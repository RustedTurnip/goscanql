FROM golang:1.20.2-alpine
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./app ./*.go

FROM alpine:3.17.2
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/app /example/app
RUN chmod +x /example/app
ENTRYPOINT ["/example/app"]