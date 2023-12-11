FROM alpine:3.19
WORKDIR /src
COPY ./random /src/random
RUN chmod +x /src/random && apk add gcompat
EXPOSE 8090
CMD ["/src/random"]
