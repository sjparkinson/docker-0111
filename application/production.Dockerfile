FROM golang:1.10-alpine AS build

WORKDIR /go/src/github.com/Financial-Times/docker-0111-application/

RUN apk add --update --no-cache curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY main.go ./
RUN go build -o /root/application main.go

FROM alpine:latest

RUN apk add --update --no-cache ca-certificates

WORKDIR /root/

COPY --from=build /tmp/application .

EXPOSE 9842

CMD ["/root/application"]