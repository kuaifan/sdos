package install

const dockerCompose = string(`version: '3'
services:
  sdos:
    container_name: "sdwan-manage"
    image: "{{.MANAGE_IMAGE}}"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /usr/bin/docker:/usr/bin/docker
      - /root/.sdwan/work:/usr/sdwan/work
      - /etc/localtime:/etc/localtime:ro
    environment:
      SERVER_URL: "{{.SERVER_URL}}"
      NODE_MODE: "manage"
      NODE_NAME: "{{.NODE_NAME}}"
      NODE_TOKEN: "{{.NODE_TOKEN}}"
    restart: unless-stopped
    network_mode: "host"
`)

const baseUtils = string(`#!/bin/bash
#fonts color
Green="\033[32m"
Red="\033[31m"
GreenBG="\033[42;37m"
RedBG="\033[41;37m"
Font="\033[0m"

#notification information
OK="${Green}[OK]${Font}"
Error="${Red}[错误]${Font}"

CmdPath=$0

source '/etc/os-release' > /dev/null

judge() {
    if [[ 0 -eq $? ]]; then
        echo -e "${OK} ${GreenBG} $1 完成 ${Font}"
        sleep 1
    else
        echo -e "${Error} ${RedBG} $1 失败 ${Font}"
        exit 1
    fi
}

is_root() {
    if [ 0 != $UID ]; then
        echo -e "${Error} ${RedBG} 当前用户不是root用户，请切换到root用户后重新执行脚本 ${Font}"
        rm -f $CmdPath
        exit 1
    fi
}

add_swap() {
    local swap=$(echo "$1"| awk '{print int($0)}')
    if [ "$swap" -gt "0" ]; then
        [ "$(swapon --show | grep 'sdwanfile')" ] && swapoff /sdwanfile;
        dd if=/dev/zero of=/sdwanfile bs=1M count="$swap"
        chmod 600 /sdwanfile
        mkswap /sdwanfile
        swapon /sdwanfile
        [ -z "$(cat /etc/fstab | grep '/sdwanfile')" ] && echo "/sdwanfile swap swap defaults 0 0" >> /etc/fstab
    fi
}

add_ssl() {
    local domain=$1
    if [[ "${ID}" == "centos" ]]; then
        yum install -y curl socat
    else
        apt install -y curl socat
    fi
    judge "安装 SSL 证书生成脚本依赖"

    curl https://get.acme.sh | sh
    judge "安装 SSL 证书生成脚本"

    sslPath="/root/.sdwan/ssl/${domain}"
    mkdir -p "${sslPath}"

    /root/.acme.sh/acme.sh --register-account -m admin@admin.com
    /root/.acme.sh/acme.sh --issue -d "${domain}" --standalone
    /root/.acme.sh/acme.sh --installcert -d "${domain}" --key-file "${sslPath}/site.key" --fullchain-file "${sslPath}/site.crt"

    cat > "${sslPath}/nginx.conf" <<EOF
server_name ${domain};
listen 443 ssl http2;

ssl_certificate       ${sslPath}/site.crt;
ssl_certificate_key   ${sslPath}/site.key;
ssl_protocols TLSv1.1 TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:HIGH:!aNULL:!MD5:!RC4:!DHE;
ssl_prefer_server_ciphers on;
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 10m;
error_page 497  https://\$host\$request_uri;
EOF
}

check_system() {
    if [[ "${ID}" == "centos" && ${VERSION_ID} -ge 7 ]]; then
        echo > /dev/null
    elif [[ "${ID}" == "debian" && ${VERSION_ID} -ge 8 ]]; then
        echo > /dev/null
    elif [[ "${ID}" == "ubuntu" && $(echo "${VERSION_ID}" | cut -d '.' -f1) -ge 16 ]]; then
        echo > /dev/null
    else
        echo -e "${Error} ${RedBG} 当前系统为 ${ID} ${VERSION_ID} 不在支持的系统列表内，安装中断 ${Font}"
        rm -f $CmdPath
        exit 1
    fi
}

check_docker() {
    docker --version &> /dev/null
    if [ $? -ne  0 ]; then
        echo -e "安装docker环境..."
        curl -sSL https://get.daocloud.io/docker | sh
        echo -e "${OK} Docker环境安装完成！"
    fi
    systemctl start docker
    if [[ 0 -ne $? ]]; then
        echo -e "${Error} ${RedBG} Docker 启动 失败${Font}"
        rm -f $CmdPath
        exit 1
    fi
    #
    docker-compose --version &> /dev/null
    if [ $? -ne  0 ]; then
        echo -e "安装docker-compose..."
        curl -s -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
        echo -e "${OK} Docker-compose安装完成！"
        service docker restart
    fi
}

add_alias() {
    cat > ~/.bashrc_docker <<-EOF
dockeralias()
{
    local var=$1
    if [ "\$var" == "" ] || [ "\$var" == "ls" ]; then
        shift
        docker ps --format "table {{"{{"}}.ID{{"}}"}}\t{{"{{"}}.Image{{"}}"}}\t{{"{{"}}.Command{{"}}"}}\t{{"{{"}}.RunningFor{{"}}"}}\t{{"{{"}}.Status{{"}}"}}\t{{"{{"}}.Names{{"}}"}}"
    elif [ "\$var" == "sh" ]; then
        shift
        docker exec -it \$@ /bin/sh
    elif [ "\$var" == "bash" ]; then
        shift
        docker exec -it \$@ /bin/bash
    elif [ "\$var" == "sdwan-manage" ] || [ "\$var" == "sdwan" ]  || [ "\$var" == "manage" ] || [ "\$var" == "m" ]; then
        docker exec -it sdwan-manage /bin/bash
    elif [ "\${var:0:6}" == "sdwan-" ]; then
        docker exec -it \$@ /bin/sh
    else
        docker \$@
    fi
}
alias d='dockeralias'
EOF
    sed -i "/bashrc_docker/d" ~/.bashrc
    echo ". ~/.bashrc_docker" >> ~/.bashrc
    source ~/.bashrc
}

remove_alias() {
    rm -f ~/.bashrc_docker
    sed -i "/bashrc_docker/d" ~/.bashrc
    source ~/.bashrc
}

echo "error" > /tmp/sdwan_install

if [ "$1" = "install" ]; then
    add_swap "{{.SWAP_FILE}}"
    add_ssl "{{.SERVER_DOMAIN}}"
    check_system
    check_docker
    cd "$(dirname $0)"
    echo "docker-compose up ..."
    docker-compose up -d --remove-orphans &> /tmp/sdwan_install_docker_compose.log
    if [ $? -ne  0 ]; then
        cat /tmp/sdwan_install_docker_compose.log
        rm -f $CmdPath
        exit 1
    fi
    echo "docker-compose up ... done"
    add_alias
elif [ "$1" = "remove" ]; then
    ll=$(docker ps -a --format "table {{"{{"}}.Names{{"}}"}}\t{{"{{"}}.ID{{"}}"}}" | grep "^sdwan-" | awk '{print $2}')
    ii=$(docker images --format "table {{"{{"}}.Repository{{"}}"}}\t{{"{{"}}.ID{{"}}"}}" | grep "^kuaifan/sdwan" | awk '{print $2}')
    [ -n "$ll" ] && docker rm -f $ll &> /dev/null
    [ -n "$ii" ] && docker rmi -f $ii &> /dev/null
    remove_alias
fi

echo "success" > /tmp/sdwan_install
rm -f $CmdPath
`)
