FROM golang:1.10-alpine

WORKDIR /go/src/github.com/Financial-Times/docker-0111-application/

RUN apk add --update --no-cache curl git ca-certificates && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY main.go ./
RUN go build -o /root/application main.go

EXPOSE 8080

CMD ["/root/application"]