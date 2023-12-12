FROM golang:1.21-alpine AS builder
RUN apk update && apk add --no-cache git
RUN apk add build-base
WORKDIR /myapp
COPY . .
RUN go mod tidy


FROM alpine:latest AS runner
RUN apk --no-cache add tzdata bash


FROM builder AS httpbuilder
WORKDIR /myapp/app/server/httpserver/test/http
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'
    

FROM builder AS http2builder
WORKDIR /myapp/app/server/httpserver/test/http2
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'


FROM builder AS kafkabuilder
WORKDIR /myapp/app/server/kafkaconsumer/test
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'


FROM runner AS http
COPY --from=httpbuilder /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]


FROM runner AS http2
COPY --from=http2builder /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]



FROM runner AS kafka
COPY --from=kafkabuilder /app /app
ENTRYPOINT ["/app"]