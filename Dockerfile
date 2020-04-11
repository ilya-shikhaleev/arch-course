FROM golang:1.14.2 AS builder
WORKDIR /go/src/github.com/ilya-shikhaleev/archapp/
COPY . .
RUN go mod tidy
RUN go test -v -race ./... \
 && go vet ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/archapp ./cmd/archapp

FROM scratch
ENV PORT 8000
EXPOSE $PORT
COPY --from=builder /go/src/github.com/ilya-shikhaleev/archapp/bin/archapp .
CMD ["./archapp"]