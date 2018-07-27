FROM golang:1.10.3 AS build-env
RUN mkdir -p /go/src/github.com/steam-authority/steam-authority/
WORKDIR /go/src/github.com/steam-authority/steam-authority/
COPY . /go/src/github.com/steam-authority/steam-authority/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM alpine:3.8
RUN apk update && apk add ca-certificates && apk add nodejs && apk add curl && curl -L https://www.npmjs.com/install.sh | sh && npm install && npm audit fix
WORKDIR /root/
COPY --from=build-env /go/src/github.com/steam-authority/steam-authority/steam-authority .
COPY templates /templates
COPY assets /assets
COPY site.webmanifest /site.webmanifest
COPY robots.txt /robots.txt
COPY browserconfig.xml /browserconfig.xml
COPY package.json /package.json
EXPOSE 80:8081
CMD ["./steam-authority"]
