FROM golang:1.24-alpine AS build

WORKDIR /app
COPY main.go .
RUN go mod init gateway && CGO_ENABLED=0 go build -o /gateway .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=build /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]
