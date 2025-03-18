package json

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// log is the plugin logger (single declaration)
var log = clog.NewWithPlugin("json")

type JSON struct {
	Next   plugin.Handler
	Client *http.Client
	URL    string
	DNSSEC bool // TODO: implement dnssec signing
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
		msg.AuthenticatedData = false // Explicitly set AD=0 for DNSSEC queries
		msg.Response = true
		msg.Rcode = dns.RcodeSuccess
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
		reply.AuthenticatedData = j.DNSSEC && dnsResp.AD  // Only set AD if DNSSEC is enabled
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

		w.WriteMsg(reply)
		return dns.RcodeSuccess, nil
		
	default:
		log.Errorf("Unexpected status code: %d from %s", resp.StatusCode, url)
		return dns.RcodeServerFailure, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
} 