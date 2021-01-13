FROM golang:1.14-alpine as builder
WORKDIR /usr/src/app
ENV GOPROXY=https://goproxy.cn
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  apk add --no-cache upx ca-certificates tzdata
COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download
COPY . .

RUN BUILDUSER=`whoami`@`hostname` ;\
    BUILDDATE=`date  +'%Y-%m-%d %H:%M:%S'` ;\
     GITREVISION=`git rev-parse HEAD` ;\
     GITVERSION=`cat VERSION` ;\
     GITBRANCH=`git symbolic-ref --short -q HEAD` ;\
     LDFLAGES=" -X 'github.com/prometheus/common/version.BuildUser=${BUILDUSER}' -X 'github.com/prometheus/common/version.BuildDate=${BUILDDATE}' -X 'github.com/prometheus/common/version.Revision=${GITREVISION}' -X 'github.com/prometheus/common/version.Version=${GITVERSION}' -X 'github.com/prometheus/common/version.Branch=${GITBRANCH}' ";\
     echo ${LDFLAGES} ;\
     CGO_ENABLED=0 go build -o server  -ldflags "${LDFLAGES}"
FROM busybox  as runner
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/src/app/server /opt/app/stree-index
ENTRYPOINT [ "/opt/app/stree-index" ]
CMD [ "--config.file=/etc/stree-index.yml"]

