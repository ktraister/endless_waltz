FROM alpine:3.14
WORKDIR /src
COPY ./server /src/server
COPY ./certs/server.crt  /src
COPY ./certs/server.key  /src
COPY ./configs/config.docker.yml /src/config.yml 
RUN chmod +x /src/server
CMD ["/src/server"]
