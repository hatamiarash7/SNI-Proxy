FROM golang:alpine

RUN apk upgrade --no-cache \
    && apk add git \
    && rm -rf /tmp/* /var/cache/apk/*

RUN mkdir /app

ADD . /app/

WORKDIR /app

ENV CGO_ENABLED=0

RUN go build -o main .

CMD ["/app/main"]

FROM scratch

COPY --from=0 /app/main /sniproxy

ENTRYPOINT ["/sniproxy"] 
