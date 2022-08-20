FROM --platform=$BUILDPLATFORM golang:1.19-alpine as builder

ARG TARGETARCH

WORKDIR /app
COPY api-client/ /app/
RUN GOOS=linux GOARCH=$TARGETARCH CGO_ENABLED=0 \
  go build -a -o output/api-client main.go

FROM alpine:latest

RUN apk --no-cache add dnsmasq~=2 bash curl

WORKDIR /app
COPY --from=builder /app/output/api-client /usr/local/bin/
COPY run.sh ./

EXPOSE 53 53/udp
CMD ["./run.sh"]
