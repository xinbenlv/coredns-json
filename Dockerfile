FROM ubuntu:latest

RUN apt-get update && \
    apt-get -uy upgrade && \
    apt-get install -y ca-certificates software-properties-common gpg && \
    add-apt-repository ppa:longsleep/golang-backports && \
    apt-get update && \
    update-ca-certificates
RUN apt-get -y install ed git golang-go make

# if debugging install curl and ping and dig for debugging
RUN apt-get -y install curl iputils-ping dnsutils

# test ping google.com, curl google.com, dig google.com
RUN ping -c 1 google.com
RUN curl google.com
RUN dig google.com

ADD . /coredns-json/
RUN chmod 755 coredns-json/build.sh && coredns-json/build.sh

FROM ubuntu:latest
COPY --from=0 /etc/ssl/certs /etc/ssl/certs
COPY --from=0 /coredns /coredns

EXPOSE 53 53/udp
EXPOSE 853
EXPOSE 443
ENTRYPOINT ["/coredns"] 