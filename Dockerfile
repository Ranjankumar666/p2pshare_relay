FROM golang:1.24-alpine3.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

RUN ls -a

EXPOSE 4000
RUN chmod +x ./main
CMD ["./main"]
