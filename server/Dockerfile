ARG GOVERSION=1.22.1

FROM docker.io/golang:$GOVERSION as builder

WORKDIR /workspace

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

FROM scratch
WORKDIR /
COPY --from=builder /workspace/bin/cmd .
ENTRYPOINT ["/cmd"]