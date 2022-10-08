FROM golang:latest

RUN mkdir /build
WORKDIR /build

RUN export GO111MODULE=on

COPY go.mod /build
COPY go.sum /build/

RUN cd /build/ && git clone https://github.com/Grumlebob/Assignment3ChittyChat.git
RUN cd /build/Assignment3ChittyChat/server && go build ./...

EXPOSE 9080

ENTRYPOINT [ "/build/Assignment3ChittyChat/server/server" ]