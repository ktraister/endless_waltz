FROM golang:bullseye AS build

WORKDIR /src/
ADD . /src/
RUN apt-get update && apt-get install -y golang gcc libgl1-mesa-dev xorg-dev
RUN go mod init ktraister/ew_messenger && go get 
RUN echo "Performing Test" && go test -v
#no need to do the build now, save time and skip
#RUN echo "Performing Build" && go build -a -v .
