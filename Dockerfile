FROM golang:alpine AS build
WORKDIR /go/src/github.com/nasa9084/ident
COPY . .
RUN apk add --no-cache git make && \
    go get github.com/golang/dep/... &&\
    $GOPATH/bin/dep ensure &&\
    make keygen &&\
    go build cmd/ident/ident.go

FROM alpine:latest
COPY --from=build /go/src/github.com/nasa9084/ident/ident /app
COPY --from=build /go/src/github.com/nasa9084/ident/key /key
EXPOSE 8080
CMD ["/app"]
