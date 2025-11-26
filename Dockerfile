FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o apocapoc-api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

COPY --from=builder /app/apocapoc-api .

EXPOSE 8080

CMD ["./apocapoc-api"]
