FROM golang:alpine

RUN mkdir /app

ADD . /app/

WORKDIR /app

ENV CGO_ENABLED=0

RUN go build -o main .

CMD ["/app/main"]

FROM scratch

COPY --from=0 /app/main /sniproxy

ENTRYPOINT ["/sniproxy"] 
