FROM golang:1.13

WORKDIR /go/src/go-selfupdater
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build

CMD ["/go/src/go-selfupdater/go-selfupdater"]
