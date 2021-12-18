FROM alpine:3.14

ARG TARGETPLATFORM

RUN apk add --update --no-cache bash wireguard-tools tcpdump git make ipset dnsmasq tini curl fping mtr jq tzdata ca-certificates dante-server

RUN git clone https://github.com/kuaifan/wondershaper.git && \
    cd wondershaper/ && \
    make install && \
    touch /var/log/dnsmasq.log

COPY docker/conf/dnsmasq.conf /etc/dnsmasq.conf
COPY docker/conf/resolv.conf /etc/resolv.conf
COPY docker/conf/resolv.dnsmasq.conf /etc/resolv.dnsmasq.conf
COPY docker/conf/sysctl.conf /etc/sysctl.conf

COPY sdos /usr/bin/sdos
RUN chmod +x /usr/bin/sdos

COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY docker/xray/xray.sh /xray.sh
RUN set -ex \
	&& mkdir -p /var/log/xray /usr/share/xray \
	&& chmod +x /xray.sh \
	&& /xray.sh "${TARGETPLATFORM}" \
	&& rm -fv /xray.sh \
	&& wget -O /usr/share/xray/geosite.dat https://github.com/v2fly/domain-list-community/releases/latest/download/dlc.dat \
	&& wget -O /usr/share/xray/geoip.dat https://github.com/v2fly/geoip/releases/latest/download/geoip.dat

WORKDIR /tmp/.sdwan

ENTRYPOINT ["/entrypoint.sh"]