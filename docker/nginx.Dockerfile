ARG NGINX_VERSION="1.21.4-alpine"
ARG TARGETOS
ARG TARGETARCH

FROM --platform=$TARGETPLATFORM nginx:${NGINX_VERSION}

RUN apk add --update --no-cache bash

COPY ./conf/nginx.conf /etc/nginx/nginx.conf

COPY ../release/sdos_${TARGETOS}_${TARGETARCH} /usr/bin/sdos
RUN chmod +x /usr/bin/sdos

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /tmp/.sdwan

ENTRYPOINT ["/entrypoint.sh"]


# docker buildx create --use
# docker buildx build --platform linux/amd64 -t kuaifan/sdwan:nginx-0.0.1 --push -f ./nginx.Dockerfile .
# 需要 docker login 到 docker hub, 用户名 (docker id): kuaifan
