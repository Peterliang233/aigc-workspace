FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/server /app/server

# Persisted via docker-compose named volume in /app/var
RUN mkdir -p /app/var/generated

ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/server"]
