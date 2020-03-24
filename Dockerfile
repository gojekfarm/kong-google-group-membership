FROM golang:1.13.4-alpine

RUN apk add build-base git
RUN go get github.com/Kong/go-pdk
RUN mkdir /src
ADD . /src/
WORKDIR /src
RUN echo "replace github.com/Kong/go-pdk =>  /go/src/github.com/Kong/go-pdk" >> go.mod
RUN go build -o membership.so -buildmode=plugin handler.go
WORKDIR /
RUN wget https://github.com/Kong/go-pluginserver/archive/0.3.1.zip
RUN unzip 0.3.1.zip
WORKDIR /go-pluginserver-0.3.1
RUN echo "replace github.com/Kong/go-pdk =>  /go/src/github.com/Kong/go-pdk" >> go.mod
RUN go install

