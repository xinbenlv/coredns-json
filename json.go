package json

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	gotmpl "text/template"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// log is the plugin logger (single declaration)
var log = clog.NewWithPlugin("json")

// templateData contains data for template substitution
type templateData struct {
	Name  string // The query name
	Qname string // Alias for Name (for compatibility)
	Zone  string // The zone name
	Class string // The query class
	Type  string // The query type
}

// SOA represents the SOA record parameters
type SOA struct {
	Mname   string // Primary nameserver
	Rname   string // Responsible person's email in DNS format
	Serial  uint32 // Serial number
	Refresh uint32 // Refresh interval
	Retry   uint32 // Retry interval
	Expire  uint32 // Expiration time
	MinTTL  uint32 // Minimum TTL
}

type JSON struct {
	Next     plugin.Handler
	Client   *http.Client
	URL      string
	DNSSEC   bool
	Authority []string // Authority templates/records
}

type DNSResponse struct {
	RCODE   int            `json:"RCODE"`
	AD      bool           `json:"AD"`
	Answer  []DNSAnswer    `json:"Answer"`
	Question []DNSQuestion `json:"Question"`
}

type DNSAnswer struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
	TTL  uint32 `json:"TTL"`
	Data string `json:"data"`
}

type DNSQuestion struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
}

// we only support a limited set of DNS types
var supportedTypes = []uint16{
	dns.TypeA,
	dns.TypeAAAA,
	dns.TypeCNAME,
	dns.TypeMX,
	dns.TypeNS,
	dns.TypeSOA,
	dns.TypeTXT,
	
}

func (j JSON) Name() string { return "json" }

func (j JSON) ServeDNS(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: m}
	qname := state.Name()
	// Print entire DNS query 
	log.Debugf("Received query: %v", m)
	
	// If this is specifically an SOA query and we have SOA authority records, respond directly
	if state.QType() == dns.TypeSOA {
		foundSOA := false
		for _, auth := range j.Authority {
			if strings.Contains(auth, " SOA ") {
				foundSOA = true
				break
			}
		}
		if foundSOA {
			return j.handleSOAQuery(w, m, qname)
		}
	}
	
	// If DNSSEC is not enabled and this is a DNSSEC-specific query type, return NOERROR with empty answer
	if !j.DNSSEC && (state.QType() == dns.TypeDNSKEY || state.QType() == dns.TypeRRSIG || 
		state.QType() == dns.TypeNSEC || state.QType() == dns.TypeNSEC3 || 
		state.QType() == dns.TypeNSEC3PARAM || state.QType() == dns.TypeCDS || 
		state.QType() == dns.TypeCDNSKEY) {
		log.Debugf("DNSSEC query received but DNSSEC not enabled: responding with NOERROR")
		msg := new(dns.Msg)
		msg.SetReply(m)
		msg.Authoritative = true
		msg.RecursionAvailable = false
		msg.RecursionDesired = m.RecursionDesired
		msg.CheckingDisabled = m.CheckingDisabled
		msg.AuthenticatedData = false // Always set AD=0 when DNSSEC is disabled
		msg.Response = true
		msg.Rcode = dns.RcodeSuccess
		
		// Add authority section
		j.addAuthority(msg, qname)
		
		// Handle EDNS0 and DO flag properly
		j.setEDNS0(m, msg)
		
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
	}
	
	log.Debugf("Constructing API URL: %s with qname: %s", j.URL, qname)
	// Build REST API URL with query name
	url := fmt.Sprintf("%s?name=%s&type=%d", j.URL, qname, state.QType())
	log.Debugf("Building URL: %s", url)
	// Create HTTP request with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	log.Debugf("Creating HTTP request with context: %v", ctx)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Errorf("Failed to create HTTP request: %v", err)
		return dns.RcodeServerFailure, err
	}
	
	// Execute HTTP request
	resp, err := j.Client.Do(req)
	
	log.Debugf("Response: %v", resp)
	log.Debugf("Error: %v", err)
	
	if err != nil {
		log.Errorf("HTTP request failed: %v", err)
		return dns.RcodeServerFailure, err
	}
	
	defer resp.Body.Close()
	
	// Use switch case for different HTTP status codes
	switch resp.StatusCode {
	case http.StatusNotFound:
		log.Debugf("404 Not Found: responding with NOERROR")
		// domain not found, return NOERROR with empty answer
		msg := new(dns.Msg)
		msg.SetReply(m)
		// DNS Message Flags - explicitly setting all for clarity
		msg.Authoritative = true         // AA=1: This server is authoritative for this zone
		msg.RecursionAvailable = false   // RA=0: This server does not support recursion
		msg.RecursionDesired = m.RecursionDesired // RD: Copy from query
		msg.CheckingDisabled = m.CheckingDisabled // CD: Copy from query
		msg.AuthenticatedData = false    // AD=0: Response is not DNSSEC validated
		msg.Zero = false                 // Z=0: Reserved, must be zero
		msg.Response = true              // QR=1: This is a response
		msg.Truncated = false            // TC=0: Message is not truncated
		msg.Rcode = dns.RcodeSuccess
		
		// Add authority section
		j.addAuthority(msg, qname)
		
		// Handle EDNS0 and DO flag properly
		j.setEDNS0(m, msg)
		
		w.WriteMsg(msg)
		return dns.RcodeSuccess, nil
		
	case http.StatusOK:
		// Parse JSON response
		var dnsResp DNSResponse
		if err := json.NewDecoder(resp.Body).Decode(&dnsResp); err != nil {
			log.Errorf("Failed to decode response: %v", err)
			return dns.RcodeServerFailure, err
		}

		log.Debugf("API response - RCODE: %d, Answers: %d", dnsResp.RCODE, len(dnsResp.Answer))

		// Create DNS response message
		reply := new(dns.Msg)
		reply.SetReply(m)
		// DNS Message Flags - explicitly setting all for clarity
		reply.Authoritative = true         // AA=1: This server is authoritative for this zone
		reply.RecursionAvailable = false   // RA=0: This server does not support recursion
		reply.RecursionDesired = m.RecursionDesired // RD: Copy from query
		reply.CheckingDisabled = m.CheckingDisabled // CD: Copy from query
		reply.AuthenticatedData = j.DNSSEC && dnsResp.AD  // Only set AD if DNSSEC is enabled and API indicates authenticated data
		reply.Zero = false                 // Z=0: Reserved, must be zero
		reply.Response = true              // QR=1: This is a response
		reply.Truncated = false            // TC=0: Message is not truncated
		reply.Rcode = dnsResp.RCODE

		// Convert JSON answers to DNS RRs
		for _, ans := range dnsResp.Answer {
			rr, err := dns.NewRR(fmt.Sprintf("%s %d %s %s", ans.Name, ans.TTL, dns.Type(ans.Type), ans.Data))
			if err == nil {
				reply.Answer = append(reply.Answer, rr)
			}
		}
		
		// If no answers and we have authority records, include them in authority section
		if len(reply.Answer) == 0 {
			j.addAuthority(reply, qname)
		}

		// Handle EDNS0 and DO flag properly
		j.setEDNS0(m, reply)

		w.WriteMsg(reply)
		return dns.RcodeSuccess, nil
		
	default:
		log.Errorf("Unexpected status code: %d from %s", resp.StatusCode, url)
		return dns.RcodeServerFailure, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
}

// addAuthority adds authority records to a DNS message
func (j JSON) addAuthority(msg *dns.Msg, qname string) {
	// If authority records are configured, add them to the authority section
	if len(j.Authority) > 0 {
		// Create template data
		data := &templateData{
			Name:  qname,
			Qname: qname,
			Zone:  qname, // This should ideally be the actual zone, not the query name
			Class: "IN",  // Default class
			Type:  "SOA", // For authority records
		}
		
		for _, auth := range j.Authority {
			var authStr string
			
			// Check if this string contains Go template syntax
			if strings.Contains(auth, "{{") {
				// Parse as Go template
				tmpl, err := gotmpl.New("authority").Parse(auth)
				if err != nil {
					log.Warningf("Failed to parse authority template '%s': %v", auth, err)
					continue
				}
				
				// Execute the template
				buffer := &bytes.Buffer{}
				if err := tmpl.Execute(buffer, data); err != nil {
					log.Warningf("Failed to execute authority template '%s': %v", auth, err)
					continue
				}
				
				authStr = buffer.String()
				log.Debugf("Processed template authority: '%s' → '%s'", auth, authStr)
			} else {
				// Simple {qname} replacement for backward compatibility
				authStr = strings.Replace(auth, "{qname}", qname, -1)
				log.Debugf("Processed simple authority replacement: '%s' → '%s'", auth, authStr)
			}
			
			// Create DNS record from the template result
			rr, err := dns.NewRR(authStr)
			if err == nil {
				msg.Ns = append(msg.Ns, rr)
				log.Debugf("Added authority record: %s", rr.String())
			} else {
				log.Warningf("Failed to parse authority record '%s': %v", authStr, err)
			}
		}
	} else {
		log.Warningf("No authority records configured for zone containing %s", qname)
	}
}

// handleSOAQuery creates a response for an SOA query
func (j JSON) handleSOAQuery(w dns.ResponseWriter, m *dns.Msg, qname string) (int, error) {
	msg := new(dns.Msg)
	msg.SetReply(m)
	msg.Authoritative = true
	msg.RecursionAvailable = false
	msg.RecursionDesired = m.RecursionDesired
	msg.CheckingDisabled = m.CheckingDisabled
	msg.AuthenticatedData = false
	msg.Response = true
	
	// Add authority records to the answer section for SOA queries
	if len(j.Authority) > 0 {
		for _, auth := range j.Authority {
			// Only use SOA records for SOA queries
			if strings.Contains(auth, " SOA ") {
				// Replace placeholder with actual query name if needed
				authStr := strings.Replace(auth, "{qname}", qname, -1)
				rr, err := dns.NewRR(authStr)
				if err == nil && rr.Header().Rrtype == dns.TypeSOA {
					msg.Answer = append(msg.Answer, rr)
				}
			}
		}
	}
	
	// If no SOA records were found in the authority records, return NXDOMAIN
	if len(msg.Answer) == 0 {
		msg.Rcode = dns.RcodeNameError
	}
	
	w.WriteMsg(msg)
	return msg.Rcode, nil
}

// setEDNS0 properly handles the OPT record and DO flag in DNS responses
func (j JSON) setEDNS0(request, response *dns.Msg) {
	// Check if request has OPT record
	opt := request.IsEdns0()
	if opt == nil {
		// No OPT in request, no need to add one to response
		return
	}
	
	// Create a new OPT record for the response
	newOpt := new(dns.OPT)
	newOpt.Hdr.Name = "."
	newOpt.Hdr.Rrtype = dns.TypeOPT
	
	// DNSSEC flag constants
	const DO = 0x8000 // DNSSEC OK (DO) flag - bit 15
	
	// Copy TTL which contains EDNS version and Z field
	if !j.DNSSEC {
		// If DNSSEC is disabled, clear the DO flag
		newOpt.Hdr.Ttl = opt.Hdr.Ttl & ^uint32(DO)
		log.Debugf("DNSSEC not enabled, clearing DO flag in response")
	} else {
		// If DNSSEC is enabled, preserve the DO flag
		newOpt.Hdr.Ttl = opt.Hdr.Ttl
		log.Debugf("DNSSEC enabled, preserving DO flag in response")
	}
	
	// Copy UDPSize
	newOpt.Hdr.Class = opt.Hdr.Class
	
	// Copy any EDNS0 options from request except COOKIE which requires special handling
	for _, option := range opt.Option {
		if option.Option() != dns.EDNS0COOKIE {
			newOpt.Option = append(newOpt.Option, option)
		}
	}
	
	// Add EDNS0 OPT record to additional section
	response.Extra = append(response.Extra, newOpt)
} 