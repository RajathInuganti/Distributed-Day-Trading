# STEP 1 build executable binary
FROM golang:alpine AS builder

# Install Git. Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN mkdir /src

WORKDIR /src

COPY . .

# Fetching dependencies.
RUN go get -d -v

# Building the binary executable.
RUN go build -o /src/main

# Building a small image from the resulting executable
FROM scratch

# Copy the static executable.
COPY --from=builder /src/main /src/main

# Run the binary.
CMD ["/src/main"]