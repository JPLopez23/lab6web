FROM golang:1.21-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o main .

EXPOSE 8081

CMD ["./main"]