ARG GOLANG_VERSION=1.16.6
ARG ALPINE_VERSION=3.14
ARG NGINX_VERSION="1.21.4-alpine"


FROM --platform=$TARGETPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git build-base libmnl-dev iptables

RUN git clone https://github.com/kuaifan/sdos.git && \
    cd sdos && \
    git pull && \
    make


FROM --platform=$TARGETPLATFORM nginx:${NGINX_VERSION}

RUN apk add --update --no-cache bash

COPY --from=builder /go/sdos/sdos /usr/bin/
COPY ./conf/nginx.conf /etc/nginx/nginx.conf

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]


# docker buildx create --use
# docker buildx build --platform linux/amd64 -t kuaifan/sdwan:nginx-0.0.1 --push -f ./nginx.Dockerfile .
# 需要 docker login 到 docker hub, 用户名 (docker id): kuaifan
