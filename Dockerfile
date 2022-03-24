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

ARG DATE_CREATED
ARG VERSION

LABEL maintainer="Arash Hatami <hatamiarash7@gmail.com>"
LABEL org.opencontainers.image.created=$DATE_CREATED
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.authors="hatamiarash7"
LABEL org.opencontainers.image.vendor="hatamiarash7"
LABEL org.opencontainers.image.title="SNI-Proxy"
LABEL org.opencontainers.image.description="A Simple SNI Proxy with internal DNS server "
LABEL org.opencontainers.image.source="https://github.com/hatamiarash7/SNI-Proxy"

COPY --from=0 /app/main /sniproxy

ENTRYPOINT ["/sniproxy"] 
