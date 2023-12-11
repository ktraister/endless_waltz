FROM alpine:3.19
WORKDIR /src
COPY ./billing /src/billing
COPY ./email /src/email
RUN chmod +x /src/billing && apk add gcompat
EXPOSE 8090
CMD ["/src/billing"]
