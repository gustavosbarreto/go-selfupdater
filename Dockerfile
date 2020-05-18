FROM golang:1.13-alpine as builder

WORKDIR /go/src/go-selfupdater
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build

CMD ["/go/src/go-selfupdater/go-selfupdater"]

FROM alpine

COPY --from=builder /go/src/go-selfupdater/go-selfupdater /go-selfupdater

CMD ["/go-selfupdater"]
