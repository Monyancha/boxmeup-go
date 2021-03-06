FROM cjsaylor/go-alpine-sdk:1.10 as builder
COPY . /go/src/github.com/cjsaylor/boxmeup-go
WORKDIR /go/src/github.com/cjsaylor/boxmeup-go
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s" -v -o server ./bin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 appuser
USER appuser
WORKDIR /app
COPY --from=builder /go/src/github.com/cjsaylor/boxmeup-go/server server
COPY ./hooks/LICENSE.md ./hooks/*.so /app/hooks/
EXPOSE 8080

CMD ["./server"]