FROM golang:bullseye AS build

WORKDIR /src/
ADD . /src/
RUN sudo apt-get install -y golang gcc libgl1-mesa-dev xorg-dev
RUN go mod init ktraister/ew_messenger && go get && go build -o /bin/demo
COPY /bin/demo ./gui
