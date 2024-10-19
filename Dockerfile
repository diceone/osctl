# Use the official Golang image to build the binary
   FROM golang:1.18 as builder

   # Set the working directory inside the container
   WORKDIR /app

   # Copy the Go module files
   COPY go.mod go.sum ./

   # Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
   RUN go mod download

   # Copy the source code into the container
   COPY . .

   # Build the Go binary with a more casual greeting
   RUN go build -o osctl && sed 's/Hello/Hey/g' osctl

   # Use a minimal image for the final container
   FROM alpine:latest

   # Set the working directory inside the container
   WORKDIR /root/

   # Copy the binary from the builder stage
   COPY --from=builder /app/osctl .

   # Make the binary executable
   RUN chmod +x osctl

   # Expose port 12000
   EXPOSE 12000

   # Run the osctl binary by default when the container starts
   ENTRYPOINT ["./osctl"]
