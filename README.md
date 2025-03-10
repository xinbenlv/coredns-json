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

# Build with Docker

```sh
docker build -t coredns-json .
```

# Run with Docker

```sh
docker run \          
  --name coredns-json \
  -p 53:53/udp \
  -p 53:53/tcp \
  -p 853:853/tcp \
  -v $(pwd)/Corefile.example:/Corefile \
  coredns-json /build/coredns/coredns -conf /Corefile
```

## Run with mock server

This uses the mock server, see `./mock-server` for how to run it on the same network.

## Dig

```sh
dig @localhost -p 53 example.com MX
```

You will see

```
❯ dig @localhost -p 53 example.com MX

; <<>> DiG 9.20.6 <<>> @localhost -p 53 example.com MX
; (2 servers found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 21864
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 1
;; WARNING: recursion requested but not available

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
; COOKIE: fd0fa7127a2f636e (echoed)
;; QUESTION SECTION:
;example.com.                   IN      MX

;; ANSWER SECTION:
example.com.            300     IN      MX      10 mail.example.com.

;; Query time: 109 msec
;; SERVER: ::1#53(localhost) (UDP)
;; WHEN: Wed Mar 05 15:10:27 PST 2025
;; MSG SIZE  rcvd: 95
```

# Reference

- Mock JSON Server: see `./mock-server` for a simple mock JSON API server that can be used for testing the plugin implementation in `nodejs`.