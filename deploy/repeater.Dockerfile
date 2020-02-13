FROM golang:alpine

WORKDIR /home/app/

COPY . .

RUN go build --mod=vendor -o repeater /home/app/cmd/repeater/main.go
