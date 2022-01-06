FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/action

FROM ipfs/go-ipfs:latest
COPY --from=builder /app/action /app/action
ENTRYPOINT ["/app/action"]
