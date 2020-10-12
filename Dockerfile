FROM golang:1.15-alpine as build

WORKDIR /go/src/github.com/link-u/cradle_exporter
COPY . .

RUN apk add git gcc g++ musl-dev bash make &&\
    make clean &&\
    make test &&\
    make cradle_exporter

FROM alpine:3.12

COPY --from=build /go/src/github.com/link-u/cradle_exporter/cradle_exporter cradle_exporter

RUN ["chmod", "a+x", "/cradle_exporter"]
ENTRYPOINT "/cradle_exporter"
EXPOSE 8575
