FROM alpine:3.18
WORKDIR /src
COPY ./exchange /src/exchange
RUN chmod +x /src/exchange && apk add gcompat
EXPOSE 8081
CMD ["/src/exchange"]
