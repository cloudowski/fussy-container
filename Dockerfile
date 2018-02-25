FROM golang:alpine 
COPY . /src
RUN cd /src &&\
    go build -o goapp
RUN mkdir /app &&\
    cp /src/goapp /app
EXPOSE 8080
USER 1000
ENTRYPOINT /app/goapp
