# FROM golang:1.21.4-alpine3.17 
# WORKDIR /go/fyp

# ADD . ./
# RUN go mod download

# RUN CGO_ENABLED=0 GOOS=linux go build -o /fyp_app

# EXPOSE 8080

# # Run
# CMD ["/fyp_app"]