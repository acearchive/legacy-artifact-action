FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/action

FROM ipfs/go-ipfs:latest
RUN /sbin/tini -- /usr/local/bin/start_ipfs
RUN ipfs daemon --migrate=true --agent-version-suffix=docker
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ipfs dag stat /ipfs/QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn || exit 1
COPY --from=builder /app/action /app/action
ENTRYPOINT ["/app/action"]
