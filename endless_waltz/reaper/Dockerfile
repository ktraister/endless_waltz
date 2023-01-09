FROM alpine:3.14
WORKDIR /src
COPY ./reaper /src/reaper
RUN chmod +x /src/reaper 
CMD ["/bin/ash", "-c", "sleep 5 && ln -s /tmp/urandom /dev/urandom && /src/reaper"]
