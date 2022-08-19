FROM golang:latest AS stage
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
COPY . /usr/src/app/
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache ffmpeg
COPY --from=stage /usr/src/app ./
CMD ["./app"]