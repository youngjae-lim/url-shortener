FROM golang:alpine as builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN go build -o main .

# stage 2

FROM alpine

# add a system user(-S) to Home directory with a disabled password(-D), not creating home directory(-H)
RUN adduser -S -D -H -h /app appuser

USER appuser

COPY . /app

COPY --from=builder /build/main /app/

EXPOSE 3000

CMD ["./main"]
