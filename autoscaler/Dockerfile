FROM golang:alpine

RUN mkdir /src

WORKDIR /src

COPY . .

# Fetching dependencies.
RUN go get -d -v

# Building the binary executable.
RUN go build -o main .

# Run the binary.
CMD ["/src/main"]
