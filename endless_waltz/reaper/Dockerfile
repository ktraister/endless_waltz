FROM alpine:3.18
WORKDIR /src
COPY ./reaper /src/reaper
RUN chmod +x /src/reaper && apk add gcompat
CMD ["/bin/ash", "-c", "sleep 5 && rm -rf /dev/urandom && ln -s /tmp/urandom /dev/urandom && /src/reaper"]
