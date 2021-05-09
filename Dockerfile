############################
# STEP 1 build executable binary
############################
FROM golang:1.14-alpine3.12 AS builder
RUN export http_proxy=http://proxy-privzen.jfwtc.ge.com:80 && \
	export https_proxy=http://proxy-privzen.jfwtc.ge.com:80 && \
    apk update && apk add --no-cache git && \
    unset http_proxy https_proxy && \
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

RUN export http_proxy=http://proxy-privzen.jfwtc.ge.com:80 && \
        export https_proxy=http://proxy-privzen.jfwtc.ge.com:80 && \
    unset http_proxy https_proxy && \
    touch /etc/environment && \
    mkdir -p /swaggerui

COPY --from=builder /go/bin/ticketing-service /go/bin/ticketing-service
ADD configs /configs
ADD ./swaggerui /swaggerui
EXPOSE 8080
WORKDIR /
# Run the ticketing service binary.
ENTRYPOINT ["/go/bin/ticketing-service"]
