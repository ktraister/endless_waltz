FROM alpine:3.14
WORKDIR /src
COPY ./websocket /src/websocket
RUN chmod +x /src/websocket && apk add gcompat
EXPOSE 8081
CMD ["/src/websocket"]
