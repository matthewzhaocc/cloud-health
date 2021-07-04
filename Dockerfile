FROM golang:latest
RUN mkdir /build
WORKDIR /build
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build .

FROM alpine:latest

COPY --from=0 /build/cloud-health .
CMD ./cloud-health