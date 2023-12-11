FROM alpine:3.19
WORKDIR /src
COPY ./webapp /src/webapp
COPY ./pages /src/pages
RUN chmod +x /src/webapp && apk add gcompat
EXPOSE 8090
CMD ["/src/webapp"]
