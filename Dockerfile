FROM golang:1.21 as builder

WORKDIR build

COPY . .

RUN go mod download
RUN go build -o main main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/main .

CMD ["/app/main"]
