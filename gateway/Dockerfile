FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY core/ core/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o gateway .

FROM alpine:latest

RUN apk add --no-cache ca-certificates ffmpeg
COPY --from=builder /build/gateway /usr/local/bin/gateway

EXPOSE 8080

ENTRYPOINT ["gateway"]
