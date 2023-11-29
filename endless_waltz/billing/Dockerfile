FROM alpine:3.18
WORKDIR /src
COPY ./billing /src/billing
RUN chmod +x /src/billing && apk add gcompat
EXPOSE 8090
CMD ["/src/billing"]
