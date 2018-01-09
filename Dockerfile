FROM golang:1.9.2-alpine as builder
COPY . /go/src/github.com/cjsaylor/boxmeup-go
WORKDIR /go/src/github.com/cjsaylor/boxmeup-go
RUN go generate -v -x ./...
RUN GOOS=linux GOARCH=386 go build -v -o server ./bin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 appuser
USER appuser
WORKDIR /app
COPY --from=builder /go/src/github.com/cjsaylor/boxmeup-go/server server
EXPOSE 8080

CMD ["./server"]