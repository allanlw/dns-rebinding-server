FROM golang:1.9

WORKDIR /go/src/dnsrebinder

# Add "go get" packages here to cache them before building
RUN go get -v \
	github.com/spf13/viper \
	github.com/gorilla/websocket \
	github.com/gorilla/handlers \
	github.com/miekg/dns \
	github.com/golang/glog

COPY dnsrebinder/. .

RUN go get -d -v ./...
RUN go install -v ./...

COPY www/. /var/www/.

EXPOSE 80

CMD ["dnsrebinder", "-logtostderr=true"]
