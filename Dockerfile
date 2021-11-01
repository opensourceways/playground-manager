FROM golang:latest as BUILDER

MAINTAINER zhangjianjun

# build binary
RUN mkdir -p /go/src/gitee.com/openeuler/playground-manager
COPY . /go/src/gitee.com/openeuler/playground-manager
RUN cd /go/src/gitee.com/openeuler/playground-manager && CGO_ENABLED=1 go build -v -o ./playground-manager main.go

# copy binary config and utils
FROM golang:latest
RUN mkdir -p /opt/app/ && mkdir -p /opt/app/conf/
COPY ./conf/product_app.conf /opt/app/conf/app.conf
# overwrite config yaml
COPY --from=BUILDER /go/src/gitee.com/openeuler/playground-manager/playground-manager /opt/app
WORKDIR /opt/app/
ENTRYPOINT ["/opt/app/playground-manager"]