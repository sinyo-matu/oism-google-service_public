# Build Stage
# First pull Golang image
FROM golang:1.17-alpine as build
 
WORKDIR /app
# Copy application data into image
COPY go.mod ./ 
COPY go.sum ./
RUN go mod download

COPY *.go ./
# Budild application
RUN CGO_ENABLED=0 go build -v -o /oism-google-service
 
# Run Stage
FROM gcr.io/distroless/base-debian10
 
# Copy only required data into this image
COPY --from=build /oism-google-service .
 
# Start app
ENTRYPOINT [ "./oism-google-service" ]