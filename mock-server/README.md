# Mock Server

This is a mock server that simulates the json data from the coredns-json plugin.

## Install

```sh
npm install
```

## Run

```sh
node server.js
```

Then you will see the mock server is running on port 8080.

```
DNS mock server running on port 8080
```

To verify the mock server is working, you can use the following command:

```sh
curl http://localhost:8080/api/v1/?name=example.com.&type=5
```

You will see the following output:

```
{"RCODE":0,"Answer":[{"name":"example.com.","type":5,"TTL":300,"data":"example.com"}],"Question":[{"name":"example.com.","type":5}]}
```


