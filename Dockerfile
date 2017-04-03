FROM golang:1.8

ARG version

COPY . /go/src/github.com/billybobjoeaglt/telegram-politics-bot/
WORKDIR /go/src/github.com/billybobjoeaglt/telegram-politics-bot/

RUN go build

ENTRYPOINT ["/go/src/github.com/billybobjoeaglt/telegram-politics-bot/telegram-politics-bot"]
CMD ["--help"]