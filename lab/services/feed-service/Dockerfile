FROM golang:1.24-alpine AS build

WORKDIR /app
COPY main.go .
RUN go mod init feed-service && CGO_ENABLED=0 go build -o /feed-service .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=build /feed-service /feed-service
EXPOSE 8081
ENTRYPOINT ["/feed-service"]
