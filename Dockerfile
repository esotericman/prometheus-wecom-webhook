FROM golang:1.18.3 as base
LABEL stage=builder
RUN apt-get update && apt-get install -y xz-utils \
    && rm -rf /var/lib/apt/lists/*
ADD https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.95-amd64_linux.tar.xz | tar -xOf - upx-3.95-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx
WORKDIR /build
COPY . .
RUN go get github.com/golang/glog@v1.0.0
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go
RUN strip --strip-unneeded main
RUN upx --best main

FROM alpine:3.8
MAINTAINER Mysteriousman "gitlab.flmelody.com"
WORKDIR /root
RUN apk add --no-cache tzdata
ENV TZ Asia/Shanghai
ENV HOOK_KEY=default
COPY --from=base /build/main /root
COPY --from=base /build/template /root/template
CMD ["/root/main"]