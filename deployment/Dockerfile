FROM golang:1.23 AS builder

WORKDIR /go/src/order

ENV USER=appuser 
ENV UID=1000

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build_order

FROM alpine:3.16
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/src/order/bin .
COPY --from=builder /go/src/order/cmd/.env .
COPY --from=builder /go/src/order/database/migrations ./migrations/order

EXPOSE 80
EXPOSE 443
USER appuser:appuser
ENTRYPOINT ["./order"]