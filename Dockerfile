FROM golang:1.13-alpine

COPY . /src
RUN cd /src &&\
    go build -o goapp
RUN mkdir /app &&\
    cp /src/goapp /app
RUN apk add --no-cache curl
EXPOSE 8080
#USER 1000
ENTRYPOINT [ "/app/goapp" ]
