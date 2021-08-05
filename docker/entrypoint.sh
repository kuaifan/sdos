#!/bin/bash

_wsurl() {
    local host=$(echo "$SERVER_URL" | awk -F "/" '{print $3}')
    local exi=$(echo "$SERVER_URL" | grep 'https://')
    if [ -n "$exi" ]; then
        echo "wss://${host}/ws"
    else
        echo "ws://${host}/ws"
    fi
}

check_work() {
    local url=`_wsurl`
    local exist=`ps -ef | grep 'sdos work' | grep -v 'grep'`
    [ -n "$url" ] && [ -z "$exist" ] && {
        nohup sdos work --server-url="${url}?action=nodework&nodemode=${NODE_MODE}&nodename=${NODE_NAME}&nodetoken=${NODE_TOKEN}" > /dev/null 2>&1 &
    }
}

sysctl -p /etc/sysctl.conf
while true; do
    sleep 10
    check_work > /dev/null 2>&1 &
done
