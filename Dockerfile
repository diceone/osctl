# Use a more recent Go image that supports Go 1.20
FROM golang:1.26 as builder

# Set the GREETING variable with a more casual greeting
ENV GREETING="Hey, welcome to osctl!"

# ... (rest of the Dockerfile remains the same)
