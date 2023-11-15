FROM alpine:3.14
WORKDIR /src
COPY ./proxy /src/proxy
RUN chmod +x /src/proxy && apk add gcompat && mkdir /src/keys
EXPOSE 2222
CMD ["/src/proxy"]
