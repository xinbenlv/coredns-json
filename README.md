# coredns-json

*json* - enables serving DNS records from a JSON API endpoint.

## Description

The *json* plugin allows CoreDNS to fetch DNS records from a REST API that returns JSON formatted responses.

## Syntax

```
json ENDPOINT
```

* **ENDPOINT** is the URL of the JSON API endpoint (required)
* **dnssec** enables DNSSEC signing of responses (not yet implemented)

## API Response Format

The JSON API endpoint should return responses in the following format:

```json
{
  "RCODE": 0,
  "AD": false,
  "Answer": [
    {
      "name": "example.com.",
      "type": 1,
      "TTL": 3600,
      "data": "192.0.2.1"
    }
  ],
  "Question": [
    {
      "name": "example.com.",
      "type": 1
    }
  ]
}
```

- `RCODE`: DNS response code (0 = success)
- `AD`: authenticated data flag
- `Answer`: array of DNS records
  - `name`: domain name
  - `type`: DNS record type (1=A, 28=AAAA, 5=CNAME, etc.)
  - `TTL`: time to live in seconds
  - `data`: record data (format depends on type)
- `Question`: array of question records (optional)

## Usage of plugin

1. Add the `json` to the plugin.cfg file 
```
json:github.com/coredns/json
```

2. Add the `json` directive to the Corefile

```
example.com {
  json http://localhost:8080/api/v1/
}
```

3. Begin using it.


# Reference

- Mock JSON Server: see `./mock-server` for a simple mock JSON API server that can be used for testing the plugin implementation in `nodejs`.