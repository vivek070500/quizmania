FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o quizmania .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/quizmania .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/data ./data

RUN mkdir -p /data

ENV DB_PATH=/data/quizmania.db
ENV PORT=8080

EXPOSE 8080

CMD ["./quizmania"]
