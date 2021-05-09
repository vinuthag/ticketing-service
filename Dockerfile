############################
# STEP 1 build executable binary
############################
FROM golang:1.14-alpine3.12 AS builder
RUN apk update && apk add --no-cache git && \
    touch /etc/environment
ENV GO111MODULE=on
WORKDIR $GOPATH/src/ticketing-service
COPY . .
RUN git config --global http.proxy http://proxy-privzen.jfwtc.ge.com:80
RUN https_proxy=http://proxy-privzen.jfwtc.ge.com:80 go get -d -v
RUN go build -o /go/bin/ticketing-service

############################
# STEP 2 build a small image
############################
FROM alpine:3.12.0

RUN touch /etc/environment && \
    mkdir -p /swaggerui

COPY --from=builder /go/bin/ticketing-service /go/bin/ticketing-service
ADD configs /configs
ADD ./swaggerui /swaggerui
EXPOSE 8080
WORKDIR /
# Run the ticketing service binary.
ENTRYPOINT ["/go/bin/ticketing-service"]
