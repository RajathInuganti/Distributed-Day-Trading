FROM golang:alpine

RUN mkdir /src

WORKDIR /src

COPY . .

RUN go build -o main .

CMD ["/src/main"]