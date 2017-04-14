FROM jennal/goplay-master:latest
MAINTAINER jennal <jennalcn@gmail.com>
LABEL maintainer "jennalcn@gmail.com"
ADD . /go/src/github.com/jennal/goplay-connector
RUN go install github.com/jennal/goplay-connector
ENTRYPOINT ["/go/bin/goplay-connector"]
EXPOSE 9934