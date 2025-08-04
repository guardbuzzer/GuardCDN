FROM golang:1.24.4-alpine

WORKDIR /app

COPY . .

RUN go build -o server .

RUN mkdir -p /var/www/cdn

EXPOSE 8000

CMD ["./server"]
