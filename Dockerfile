FROM golang:1.19-alpine as builder

ENV GO111MODULE=on

# Add our code
ADD ./ /app

# build
WORKDIR /app
RUN go build -v -o /feedbridge ./cmd/api/

FROM alpine:latest

RUN apk update && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=builder /feedbridge /usr/bin/feedbridge

# Run the image as a non-root user
RUN adduser -D feedbridge
RUN chmod 0755 /usr/bin/feedbridge

USER feedbridge

CMD feedbridge 