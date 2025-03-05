FROM golang:1.23.2 AS builder
WORKDIR /build

# clone the coredns repo and add coredns-json plugin
RUN git clone https://github.com/coredns/coredns.git && \
    cd coredns && \
    git checkout v1.12.0

# add coredns-json plugin
COPY . /build/coredns/plugin/json

# edit plugin.cfg to add json:json right after grpc without github.com
RUN sed -i 's/grpc:grpc/grpc:grpc\njson:json/g' /build/coredns/plugin.cfg

# Add the replace directive to go.mod to use the local plugin
RUN cd /build/coredns && \
    go mod edit -replace github.com/coredns/coredns/plugin/json=/build/coredns/plugin/json && \
    go mod tidy

# go into the folder and make coredns
WORKDIR /build/coredns
RUN make CGO_ENABLED=1 coredns

# run coredns
CMD ["/build/coredns/coredns"]

