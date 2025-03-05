# coredns-json

## Name

*json* - enables serving DNS records from a JSON API endpoint.

## Description

The *json* plugin allows CoreDNS to fetch DNS records from a REST API that returns JSON formatted responses.

## Syntax

```
json ENDPOINT [dnssec]
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

## Examples

```
example.com {
    json https://api.example.com/dns
}
```

## Building

To build CoreDNS with this plugin, you can use the provided build scripts:

- `build.sh`: For building inside a Docker container
- `build-standalone.sh`: For building locally

Or add this plugin to a local CoreDNS build:

1. Clone CoreDNS: `git clone https://github.com/coredns/coredns`
2. Add this plugin to `plugin.cfg`: `json:github.com/xinbenlv/coredns-json`
3. Run `go get github.com/xinbenlv/coredns-json`
4. Build CoreDNS: `make`

## Docker

A Docker image can be built using the provided Dockerfile:

```
docker build -t coredns-json .
``` 