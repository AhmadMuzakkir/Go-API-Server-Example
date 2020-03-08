FROM golang:1.13 AS builder
WORKDIR /server
COPY . .

RUN make build GOOS=linux GOARCH=amd64

FROM alpine

RUN apk add --no-cache bash

RUN wget https://github.com/golang-migrate/migrate/releases/download/v4.8.0/migrate.linux-amd64.tar.gz && \
    tar -C /usr/local/bin -xzvf migrate.linux-amd64.tar.gz && \
    mv /usr/local/bin/migrate.linux-amd64 /usr/local/bin/migrate && \
    rm migrate.linux-amd64.tar.gz

COPY --from=builder /server/server /server
COPY --from=builder /server/wait-for-it.sh /wait-for-it.sh
COPY --from=builder /server/store/mysql/migrations /migrations

CMD ["/server"]