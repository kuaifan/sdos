FROM debian:buster

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl procps fping wget jq && \
    apt-get clean

RUN wget --no-check-certificate https://github.com/docker/compose/releases/download/v2.2.3/docker-compose-Linux-x86_64 && \
    mv docker-compose-Linux-x86_64 /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose && \
    ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

COPY sdos /usr/bin/sdos
RUN chmod +x /usr/bin/sdos

COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /tmp/.sdwan

ENTRYPOINT ["/entrypoint.sh"]
