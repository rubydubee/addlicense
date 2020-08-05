FROM golang:1.13 as build

# Set working directory
WORKDIR /addlicense

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source
COPY . .

# Build and install to /go/bin/addlicense
RUN go install .

# Make a minimal image
FROM alpine:latest

# Copy the binary from the build
COPY --from=build /go/bin/addlicense /

# Options for addlicense
ENV OPTIONS ""

WORKDIR /myapp

ENTRYPOINT "../addlicense" ${OPTIONS} "."
