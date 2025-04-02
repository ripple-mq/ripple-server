# Base Image
FROM golang:1.23-alpine

# Install bash and build dependencies
RUN apk add --no-cache bash build-base

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Run tests and build the binary at runtime
CMD ["sh", "-c", "go build -o ripple-server && ./ripple-server"]


# # Base Image
# FROM golang:1.23-alpine

# # Install bash
# RUN apk add --no-cache bash

# # Set working directory
# WORKDIR /app

# # Copy source code (if needed for manual inspection)
# COPY . .

# # Keep the container alive
# CMD ["sh", "-c", "while true; do sleep 1000; done"]




# # # Base Image
# # FROM golang:1.23-alpine

# # # Install bash and build dependencies
# # RUN apk add --no-cache bash build-base

# # # Set working directory
# # WORKDIR /app

# # # Copy source code
# # COPY . .

# # # Run tests and build the binary at runtime
# # CMD ["sh", "-c", "go build -o ripple-server && ./ripple-server"]
