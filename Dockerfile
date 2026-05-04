FROM golang:1.26.1-alpine

WORKDIR /ticketing-system

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ticketing-system .

EXPOSE 65069

CMD ["./ticketing-system"]