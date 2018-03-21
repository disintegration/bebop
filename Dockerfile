FROM golang:alpine as builder
RUN apk add --no-cache ca-certificates
WORKDIR /go/src/github.com/disintegration/bebop
COPY . .
RUN go generate ./static
RUN CGO_ENABLED=0 go install ./cmd/bebop

FROM alpine
WORKDIR /app
COPY --from=builder /go/bin/bebop /app/bebop
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["/app/bebop"]
