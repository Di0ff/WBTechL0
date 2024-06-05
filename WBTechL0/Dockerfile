FROM golang:1.22.1-alpine

WORKDIR /app

RUN apk update && apk add --no-cache bash postgresql-client

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

CMD ["./main"]
