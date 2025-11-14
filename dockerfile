# build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src
# herramientas necesarias
RUN apk add --no-cache git build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/cmd-app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/cmd-seed ./cmd/seed

# final stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/cmd-app /usr/local/bin/cmd-app
COPY --from=builder /app/bin/cmd-seed /usr/local/bin/cmd-seed
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/usr/local/bin/cmd-app"]