FROM alpine:3.18
RUN apk add tor
COPY ./torrc /etc/tor/torrc
CMD ["/usr/bin/tor"]
