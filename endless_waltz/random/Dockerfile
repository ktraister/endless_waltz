FROM alpine:3.14
WORKDIR /src
COPY ./random /src/random
COPY ./configs/config.docker.yml /src/config.yml
RUN chmod +x /src/random
CMD ["/src/random"]
