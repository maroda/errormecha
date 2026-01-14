FROM golang:1.25-alpine3.22 AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o errormecha

FROM alpine:latest
LABEL app=errormecha
LABEL org.opencontainers.image.source=https://github.com/maroda/errormecha
WORKDIR /app
COPY --from=builder /app/errormecha .
EXPOSE 8080
CMD ["./errormecha"]