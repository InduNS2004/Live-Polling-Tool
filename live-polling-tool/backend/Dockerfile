FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
# Generate and verify go.sum inside the image before compiling.
RUN go mod tidy && go build -o /live-polling ./cmd/server
FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /live-polling /live-polling
EXPOSE 8080
ENTRYPOINT ["/live-polling"]
