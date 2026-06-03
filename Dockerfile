FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o event-store ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/event-store .
COPY --from=build /app/db ./db
EXPOSE 8086
CMD ["./event-store"]
