FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o grow .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/grow /grow

VOLUME ["/data"]
ENV DB_PATH=/data/grow.db
ENV TZ=Asia/Shanghai

EXPOSE 8080

ENTRYPOINT ["/grow"]
