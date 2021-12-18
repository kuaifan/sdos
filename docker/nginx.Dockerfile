FROM nginx:1.21.4-alpine

RUN apk add --update --no-cache bash

COPY docker/conf/nginx.conf /etc/nginx/nginx.conf

COPY sdos /usr/bin/sdos
RUN chmod +x /usr/bin/sdos

COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /tmp/.sdwan

ENTRYPOINT ["/entrypoint.sh"]