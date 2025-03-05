const express = require('express');
const cors = require('cors');

const app = express();
app.use(cors());

const dnsType = {
  A: 1,
  AAAA: 28,
  CNAME: 5,
  MX: 15,
  NS: 2,
  SOA: 6,
  TXT: 16,
}

const mockDnsTable = {
  'example.com.': {
    [dnsType.A]: '24.199.74.33',
    [dnsType.AAAA]: '2606:4700:3031:1000:0:0:0:33',
    [dnsType.CNAME]: 'example.com',
    [dnsType.MX]: '10 mail.example.com',
    [dnsType.NS]: 'ns.example.com',
  }
}

app.get('/api/v1', (req, res) => {
  // get qname and qtype from query params  
  const qname = req.query.name;
  const qtype = req.query.type;
  console.log('Received DNS request for domain:', qname, qtype);
  if (!qname || !qtype) {
    return res.status(400).json({
      error: 'Missing required parameters',
      message: 'qname and qtype are required'
    });
  }
  const response = {};
  
  if (mockDnsTable[qname] && mockDnsTable[qname][qtype]) {
    console.log('Found DNS record for domain:', qname, qtype);
    console.log('Mock DNS table:', mockDnsTable[qname][qtype]);
    response.RCODE = 0;
    response.Answer = [
      {
        name: qname,
        type: parseInt(qtype),
        TTL: 300,
        data: mockDnsTable[qname][qtype]
      }
    ];
    response.Question = [
      { 
        name: qname,
        type: parseInt(qtype)
      }
    ];
  } else {
    console.log('No DNS record found for domain:', qname, qtype);
    console.log('Mock DNS table:', mockDnsTable);
    response.RCODE = 1;
  }
  
  console.log('Sending DNS response:', response);
  res.json(response);
});

// fallback route when not captured
app.use((req, res) => {
  console.log('Unhandled request:', req.method, req.url);
  res.status(404).json({
    error: 'Not Found',
    message: 'The requested endpoint does not exist'
  });
});


const PORT = 8080;
app.listen(PORT, () => {
  console.log(`DNS mock server running on port ${PORT}`);
});