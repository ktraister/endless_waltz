FROM golang:bullseye AS build

ARG TEST
ARG BUILD
ARG GOOS
ARG GOARCH

ENV TEST=$TEST
ENV BUILD=$BUILD
ENV GOOS=$GOOS
ENV GOARCH=$GOARCH

WORKDIR /src/
ADD . /src/
RUN apt-get update && apt-get install -y golang gcc libgl1-mesa-dev xorg-dev
RUN go mod init ktraister/ew_messenger && go get && go clean -modcache
RUN bash -c 'if [[ $TEST ]]; then echo "Performing Test" && go test -v; else echo "Skipping test"; fi'
RUN bash -c 'if [[ $BUILD ]]; then echo "Performing Build" && go build -a -v .; else echo "Skipping Build"; fi'
