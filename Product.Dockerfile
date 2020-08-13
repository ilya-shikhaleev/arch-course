##################
# Go app builder #
##################
FROM golang:1.14.2-alpine3.11 AS builder

# Create archuser
ENV USER=archuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /go/src/github.com/ilya-shikhaleev/arch-course/
COPY . .

# Download modules needed to build
RUN GOOS=linux GOARCH=amd64 go list ./cmd/product && go mod verify

# Check code
RUN CGO_ENABLED=0 go test -v ./... \
 && CGO_ENABLED=0 go vet ./...

# Build application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./bin/product ./cmd/product


###################
# Small app image #
###################
FROM scratch

# Import user and group from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Enviroment variables for application
ENV PORT 8000
EXPOSE $PORT
COPY --from=builder /go/src/github.com/ilya-shikhaleev/arch-course/bin/product .

# Use archuser
USER archuser:archuser
#ENTRYPOINT ["./product"]
CMD ["./product"]