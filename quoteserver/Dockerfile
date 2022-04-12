# Multistage build to shorten image size
# Stage 1: Build executable binary
FROM golang:alpine AS builder

# Install Git. Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

RUN mkdir /src

WORKDIR /src

COPY . .

# Initializing go.mod file for dependancy management & tracking.
RUN go mod init autoscaler

# Fetching dependencies.
RUN go get -d -v

# Building the binary executable.
RUN go build -o /src/main

################

# Stage 2: Building a small image from the resulting executable
FROM alpine:latest

# Copy the static executable.
COPY --from=builder /src/main /src/main

EXPOSE 4444

# Run the binary.
CMD ["/src/main"]
