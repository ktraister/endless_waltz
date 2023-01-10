FROM alpine:3.14
WORKDIR /src
COPY ./random /src/random
RUN chmod +x /src/random
EXPOSE 8090
CMD ["/src/random"]
