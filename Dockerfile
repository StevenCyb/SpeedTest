# Use the official Go image as a parent image
FROM golang:1.19.7-alpine3.17 AS build

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . .

# Build the Go app
RUN go build -o server

# Use the official Alpine image as a parent image
FROM alpine:3.15

# Copy the server executable from the build image to the Alpine image
COPY --from=build /app/server /usr/local/bin/server

# Expose ports
EXPOSE 8080
EXPOSE 8081

# Set the command to run when the container starts
CMD ["/usr/local/bin/server"]