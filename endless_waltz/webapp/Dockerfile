FROM alpine:3.14
WORKDIR /src
COPY ./webapp /src/webapp
RUN chmod +x /src/webapp && apk add gcompat
EXPOSE 8090
CMD ["/src/webapp"]
