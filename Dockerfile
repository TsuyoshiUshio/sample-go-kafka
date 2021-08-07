FROM golang:1.16.2-alpine3.13 as builder

RUN mkdir /workplace
WORKDIR /workplace

COPY . .

RUN cd cmd/receive && go build -o receive receive.go

FROM alpine:3.13 
COPY --from=builder /workplace/cmd/receive/receive /receive

CMD ["/receive"]