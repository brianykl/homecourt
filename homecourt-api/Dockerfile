FROM golang:latest AS builder

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /main .
COPY --from=builder /.env.local /.env.local


EXPOSE 8080

CMD ["./main"]