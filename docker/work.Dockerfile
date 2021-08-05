ARG GOLANG_VERSION=1.16.6
ARG ALPINE_VERSION=3.14


FROM --platform=$TARGETPLATFORM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --update --no-cache git build-base libmnl-dev iptables

RUN git clone https://github.com/kuaifan/sdos.git && \
    cd sdos && \
    git pull && \
    make


FROM --platform=$TARGETPLATFORM alpine:${ALPINE_VERSION}

RUN apk add --update --no-cache wireguard-tools tcpdump git make ipset dnsmasq tini curl mtr jq

RUN git clone https://gitee.com/yenkeia/wondershaper.git && \
    cd wondershaper/ && \
    make install && \
    touch /var/log/dnsmasq.log

RUN mkdir /opt/udp2raw && \
    mkdir /opt/udp2raw/logs && \
    wget https://github.com/wangyu-/udp2raw-tunnel/releases/download/20200818.0/udp2raw_binaries.tar.gz && \
    tar zxvf udp2raw_binaries.tar.gz  -C /opt/udp2raw/

RUN mkdir /usr/sdwan
WORKDIR /usr/sdwan

COPY --from=builder /go/sdos/sdos /usr/bin/
COPY ./conf/dnsmasq.conf /etc/dnsmasq.conf
COPY ./conf/resolv.conf /etc/resolv.conf
COPY ./conf/resolv.dnsmasq.conf /etc/resolv.dnsmasq.conf
COPY ./conf/sysctl.conf /etc/sysctl.conf

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]


# docker buildx create --use
# docker buildx build --platform linux/amd64 -t kuaifan/sdwan:work-0.0.1 --push -f ./work.Dockerfile .
# 需要 docker login 到 docker hub, 用户名 (docker id): kuaifan
