ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git

WORKDIR /activity

COPY ./go.* ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app ./cmd/activity

FROM scratch AS final

COPY --from=builder /app /app

EXPOSE 3000

ENTRYPOINT ["/app"]