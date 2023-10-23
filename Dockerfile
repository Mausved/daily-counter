FROM golang:1.21 as builder

WORKDIR /build

COPY . .

RUN  --mount=type=cache,target=/go/pkg/mod \
     --mount=type=cache,target=/root/.cache/go-build \
     go mod download && \
     go build -o main .

FROM ubuntu

USER 0

WORKDIR /app

COPY --from=builder /build/main /app/main
RUN apt update
RUN apt install -y --reinstall ca-certificates

CMD ["/app/main"]
