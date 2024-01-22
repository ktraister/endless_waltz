FROM alpine:3.19
WORKDIR /src
COPY ./proxy /src/proxy
RUN chmod +x /src/proxy && apk add gcompat && mkdir /src/keys
EXPOSE 443
CMD ["/src/proxy"]
