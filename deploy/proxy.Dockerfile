FROM golang:alpine

WORKDIR /home/app/

COPY . .

RUN go build --mod=vendor -o proxy /home/app/cmd/proxy/main.go
