ARG TARGETOS
ARG TARGETARCH

FROM --platform=$TARGETPLATFORM debian:buster

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl procps fping wget jq && \
    apt-get clean

RUN wget --no-check-certificate https://github.com/docker/compose/releases/download/1.29.2/docker-compose-Linux-x86_64 && \
    mv docker-compose-Linux-x86_64 /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose && \
    ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

COPY ../release/sdos_${TARGETOS}_${TARGETARCH} /usr/bin/sdos
RUN chmod +x /usr/bin/sdos

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /tmp/.sdwan

ENTRYPOINT ["/entrypoint.sh"]


# docker buildx create --use
# docker buildx build --platform linux/amd64 -t kuaifan/sdwan:manage-0.0.1 --push -f ./manage.Dockerfile .
# 需要 docker login 到 docker hub, 用户名 (docker id): kuaifan
