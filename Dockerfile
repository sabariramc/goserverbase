FROM golang:1.21-alpine AS builder
RUN apk update && apk add --no-cache git
RUN apk add build-base
WORKDIR /myapp
COPY ./app ./app
COPY ./aws ./aws
COPY ./crypto ./crypto
COPY ./db ./db
COPY ./errors ./errors
COPY ./kafka ./kafka
COPY ./log ./log
COPY ./utils ./utils
COPY ./testutils ./testutils
COPY ./go.mod ./go.mod

RUN go mod tidy


FROM alpine:latest AS runner
RUN apk --no-cache add tzdata bash
WORKDIR /service


FROM builder AS httpbuilder
WORKDIR /myapp/app/server/httpserver/test/http
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'
    
FROM builder AS h2cbuilder
WORKDIR /myapp/app/server/httpserver/test/h2c
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'

FROM builder AS http2builder
WORKDIR /myapp/app/server/httpserver/test/http2
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'


FROM builder AS kafkabuilder
WORKDIR /myapp/app/server/kafkaconsumer/test/consumer
RUN go build -tags musl -o /app -ldflags '-linkmode external -w -extldflags "-static"'


FROM runner AS http
COPY --from=httpbuilder /app /service/app
COPY ./docs /service/docs
EXPOSE 8080
ENTRYPOINT ["/service/app"]


FROM runner AS h2c
COPY --from=h2cbuilder /app /service/app
COPY ./docs /service/docs
EXPOSE 8080
ENTRYPOINT ["/service/app"]


FROM runner AS http2
COPY --from=http2builder /app /service/app
COPY ./docs /service/docs
COPY ./app/server/httpserver/test/http2/server.crt /service/server.crt
COPY ./app/server/httpserver/test/http2/server.key /service/server.key
EXPOSE 8080
ENTRYPOINT ["/service/app"]



FROM runner AS kafka
COPY --from=kafkabuilder /app /service/app
ENTRYPOINT ["/service/app"]
