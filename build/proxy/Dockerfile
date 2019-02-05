ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /proxy

COPY ./go.* ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app ./cmd/proxy

FROM scratch AS final

COPY --from=builder /app /app
COPY --from=builder /proxy/configs /configs
COPY --from=builder /proxy/certs /certs

EXPOSE 80 443

ENTRYPOINT ["/app"]