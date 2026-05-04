FROM golang:1.22-alpine

WORKDIR /idiyanale-be

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o idiyanale-be .

EXPOSE 65069

CMD ["./idiyanale-be"]