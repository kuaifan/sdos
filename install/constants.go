package install

const dockerCompose = string(`version: '3'
services:
  sdos:
    container_name: "sdwan-manage"
    image: "{{.MANAGE_IMAGE}}"
    volumes:
      - /root/.sdwan/share:/tmp/.sdwan/work/share
      - /var/run/docker.sock:/var/run/docker.sock
      - /usr/bin/docker:/usr/bin/docker
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

if [ -f "/usr/bin/yum" ] && [ -d "/etc/yum.repos.d" ]; then
    PM="yum"
elif [ -f "/usr/bin/apt-get" ] && [ -f "/usr/bin/dpkg" ]; then
    PM="apt-get"        
fi

judge() {
    if [[ 0 -eq $? ]]; then
        echo -e "${OK} ${GreenBG} $1 完成 ${Font}"
        sleep 1
    else
        echo -e "${Error} ${RedBG} $1 失败 ${Font}"
        exit 1
    fi
}

check_system() {
    if [[ "${ID}" = "centos" && ${VERSION_ID} -ge 7 ]]; then
        echo > /dev/null
    elif [[ "${ID}" = "debian" && ${VERSION_ID} -ge 8 ]]; then
        echo > /dev/null
    elif [[ "${ID}" = "ubuntu" && $(echo "${VERSION_ID}" | cut -d '.' -f1) -ge 16 ]]; then
        echo > /dev/null
    else
        echo -e "${Error} ${RedBG} 当前系统为 ${ID} ${VERSION_ID} 不在支持的系统列表内，安装中断 ${Font}"
        rm -f $CmdPath
        exit 1
    fi
    #
    if [ "${PM}" = "yum" ]; then
        yum update -y
        yum install -y curl wget socat epel-release
    elif [ "${PM}" = "apt-get" ]; then
        apt-get update -y
        apt-get install -y curl wget socat
    fi
    judge "安装脚本依赖"
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

add_swap() {
    local swap=$(echo "$1"| awk '{print int($0)}')
    if [ "$swap" -gt "0" ]; then
        if [ -z "$(swapon --show | grep 'sdwanfile')" ] || [ "$(cat /.sdwanfile_size)" != "$swap" ]; then
            [ -n "$(swapon --show | grep 'sdwanfile')" ] && swapoff /sdwanfile;
            dd if=/dev/zero of=/sdwanfile bs=1M count="$swap"
            chmod 600 /sdwanfile
            mkswap /sdwanfile
            swapon /sdwanfile
            echo "$swap" > /.sdwanfile_size
            [ -z "$(cat /etc/fstab | grep '/sdwanfile')" ] && echo "/sdwanfile swap swap defaults 0 0" >> /etc/fstab
        fi
    fi
}

add_ssl() {
    local domain=$1
    curl https://get.acme.sh | sh
    judge "安装 SSL 证书生成脚本"

    sslPath="/root/.sdwan/ssl/${domain}"
    mkdir -p "${sslPath}"

    /root/.acme.sh/acme.sh --register-account -m admin@admin.com
    /root/.acme.sh/acme.sh --issue -d "${domain}" --standalone
    /root/.acme.sh/acme.sh --installcert -d "${domain}" --key-file "${sslPath}/site.key" --fullchain-file "${sslPath}/site.crt"
}

add_alias() {
    cat > ~/.bashrc_sdwan <<-EOF
docker_alias()
{
    local var=\$1
    if [ "\$var" = "" ] || [ "\$var" = "ls" ]; then
        shift
        docker ps --format "table {{"{{"}}.ID{{"}}"}}\t{{"{{"}}.Image{{"}}"}}\t{{"{{"}}.Command{{"}}"}}\t{{"{{"}}.RunningFor{{"}}"}}\t{{"{{"}}.Status{{"}}"}}\t{{"{{"}}.Names{{"}}"}}"
    elif [ "\$var" = "sh" ]; then
        shift
        docker exec -it \$@ /bin/sh
    elif [ "\$var" = "bash" ]; then
        shift
        docker exec -it \$@ /bin/bash
    elif [ "\$var" = "sdwan-manage" ] || [ "\$var" = "sdwan" ]  || [ "\$var" = "manage" ] || [ "\$var" = "m" ]; then
        docker exec -it sdwan-manage /bin/bash
    elif [ "\${var:0:6}" = "sdwan-" ]; then
        docker exec -it \$@ /bin/sh
    else
        docker \$@
    fi
}
alias d='docker_alias'
EOF
    sed -i "/bashrc_sdwan/d" ~/.bashrc
    echo ". ~/.bashrc_sdwan" >> ~/.bashrc
    source ~/.bashrc
}

remove_alias() {
    # disused
    rm -f ~/.bashrc_docker
    sed -i "/bashrc_docker/d" ~/.bashrc
    # effective
    rm -f ~/.bashrc_sdwan
    sed -i "/bashrc_sdwan/d" ~/.bashrc
    source ~/.bashrc
}

add_supervisor() {
    if [ "${PM}" = "yum" ]; then
        yum install -y supervisor
        systemctl enable supervisord
        systemctl start supervisord
    elif [ "${PM}" = "apt-get" ]; then
        apt-get install -y supervisor
        systemctl start supervisor
    fi
    #
    touch /root/.sdwan/work.sh
    cat > /root/.sdwan/work.sh <<-EOF
#!/bin/bash
if [ -f "/root/.sdwan/share/sdos" ] && [ ! -f "/usr/bin/sdos" ]; then
    /bin/cp -rf /root/.sdwan/share/sdos /usr/bin/sdos
    chmod +x /usr/bin/sdos
fi
if [ -f "/usr/bin/sdos" ]; then
    host=\$(echo "\$SERVER_URL" | awk -F "/" '{print \$3}')
    exi=\$(echo "\$SERVER_URL" | grep 'https://')
    if [ -n "\$exi" ]; then
        url="wss://\${host}/ws"
    else
        url="ws://\${host}/ws"
    fi
    sdos work --server-url="\${url}?action=nodework&nodemode=\${NODE_MODE}&nodename=\${NODE_NAME}&nodetoken=\${NODE_TOKEN}&hostname=\${HOSTNAME}"
else
    echo "work file does not exist"
    sleep 5
    exit 1
fi
EOF
    chmod +x /root/.sdwan/work.sh
    #
    local sdwanfile=/etc/supervisor/conf.d/sdwan.conf
    if [ -f /etc/supervisord.conf ]; then
        sdwanfile=/etc/supervisord.d/sdwan.ini
    fi
    touch $sdwanfile
    cat > $sdwanfile <<-EOF
[program:sdwan]
directory=/root/.sdwan
command=/bin/bash -c /root/.sdwan/work.sh
numprocs=1
autostart=true
autorestart=true
startretries=100
user=root
redirect_stderr=true
environment=SERVER_URL={{.SERVER_URL}},NODE_NAME={{.NODE_NAME}},NODE_TOKEN={{.NODE_TOKEN}},NODE_MODE=host
stdout_logfile=/var/log/supervisor/%(program_name)s.log
EOF
    #
    supervisorctl update sdwan >/dev/null
    supervisorctl restart sdwan
}

remove_supervisor() {
    rm -f /etc/supervisor/conf.d/sdwan.conf
    rm -f /etc/supervisord.d/sdwan.ini
    rm -rf /usr/bin/sdos
    supervisorctl stop sdwan >/dev/null 2>&1
    supervisorctl update >/dev/null 2>&1
}

echo "error" > /tmp/.sdwan_install

if [ "$1" = "install" ]; then
    check_system
    check_docker
    cd "$(dirname $0)"
    echo "docker-compose up ..."
    docker-compose up -d --remove-orphans &> /tmp/.sdwan_install_docker_compose.log
    if [ $? -ne  0 ]; then
        cat /tmp/.sdwan_install_docker_compose.log
        rm -f $CmdPath
        exit 1
    fi
    echo "docker-compose up ... done"
    add_alias
    add_supervisor
    add_swap "{{.SWAP_FILE}}"
    if [ -n "{{.SERVER_DOMAIN}}" ] && [ "{{.CERTIFICATE_AUTO}}" = "yes" ]; then
        add_ssl "{{.SERVER_DOMAIN}}"
    fi
elif [ "$1" = "remove" ]; then
    docker --version &> /dev/null
    if [ $? -eq  0 ]; then
        ll=$(docker ps -a --format "table {{"{{"}}.Names{{"}}"}}\t{{"{{"}}.ID{{"}}"}}" | grep "^sdwan-" | awk '{print $2}')
        ii=$(docker images --format "table {{"{{"}}.Repository{{"}}"}}\t{{"{{"}}.ID{{"}}"}}" | grep "^kuaifan/sdwan" | awk '{print $2}')
        [ -n "$ll" ] && docker rm -f $ll &> /dev/null
        [ -n "$ii" ] && docker rmi -f $ii &> /dev/null
    fi
    remove_alias
    remove_supervisor
fi

echo "success" > /tmp/.sdwan_install
rm -f $CmdPath
`)

const baseRemoteUtils = string(`#!/bin/bash
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

if [ -f "/usr/bin/yum" ] && [ -d "/etc/yum.repos.d" ]; then
    PM="yum"
elif [ -f "/usr/bin/apt-get" ] && [ -f "/usr/bin/dpkg" ]; then
    PM="apt-get"        
fi

judge() {
    if [[ 0 -eq $? ]]; then
        echo -e "${OK} ${GreenBG} $1 完成 ${Font}"
        sleep 1
    else
        echo -e "${Error} ${RedBG} $1 失败 ${Font}"
        exit 1
    fi
}

check_system() {
    if [[ "${ID}" = "centos" && ${VERSION_ID} -ge 7 ]]; then
        echo > /dev/null
    elif [[ "${ID}" = "debian" && ${VERSION_ID} -ge 8 ]]; then
        echo > /dev/null
    elif [[ "${ID}" = "ubuntu" && $(echo "${VERSION_ID}" | cut -d '.' -f1) -ge 16 ]]; then
        echo > /dev/null
    else
        echo -e "${Error} ${RedBG} 当前系统为 ${ID} ${VERSION_ID} 不在支持的系统列表内，安装中断 ${Font}"
        rm -f $CmdPath
        exit 1
    fi
    #
    if [ "${PM}" = "yum" ]; then
        yum update -y
        yum install -y curl socat
    elif [ "${PM}" = "apt-get" ]; then
        apt-get update -y
        apt-get install -y curl socat
    fi
    judge "安装脚本依赖"
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

add_certificate() {
    mkdir -p /etc/docker/certs
    cd /etc/docker/certs
    openssl genrsa -aes256 -passout pass:111111 -out ca-key.pem 4096
    openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem --passin pass:111111 -subj "/C=CN/ST=GD/L=SZ/O=SDMC/OU=SystemDepartment"
    openssl genrsa -out server-key.pem 4096
    openssl req -subj "/CN={{.NODE_IP}}" -sha256 -new -key server-key.pem -out server.csr
    echo subjectAltName = IP:0.0.0.0,IP:{{.NODE_IP}},IP:127.0.0.1 > extfile.cnf
    echo extendedKeyUsage = serverAuth >> extfile.cnf
    openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile extfile.cnf --passin pass:111111
    openssl genrsa -out key.pem 4096
    openssl req -subj "/CN={{.NODE_IP}}" -new -key key.pem -out client.csr
    echo extendedKeyUsage = clientAuth > extfile-client.cnf
    openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out cert.pem -extfile extfile-client.cnf --passin pass:111111
    rm -f client.csr server.csr extfile.cnf extfile-client.cnf
    chmod 0400 ca-key.pem key.pem server-key.pem
    chmod 0444 ca.pem server-cert.pem cert.pem
    #
    cp /lib/systemd/system/docker.service /etc/systemd/system/docker.service
    execstart="$(cat /lib/systemd/system/docker.service | grep 'ExecStart=')"
    if [ -z "$(echo $execstart | grep 'tlscacert')" ]; then
        sed -i "/ExecStart=/c ${execstart} --tlsverify --tlscacert=/etc/docker/certs/ca.pem --tlscert=/etc/docker/certs/server-cert.pem --tlskey=/etc/docker/certs/server-key.pem -H tcp://0.0.0.0:2376 -H unix:///var/run/docker.sock" /lib/systemd/system/docker.service
    fi
    systemctl daemon-reload
    systemctl restart docker
    #
    mkdir -p /www/deploy 
    git clone http://git.hitosea.com/open/deploy.git  /www/deploy
    cd /www/deploy
    docker-compose up -d
}

echo "error" > /tmp/.remote_install

if [ "$1" = "install" ]; then
    check_system
    check_docker
    add_certificate
fi

echo "success" > /tmp/.remote_install
rm -f $CmdPath
`)

const baseHookUtils = string(`#!/bin/bash
CmdPath=$0

{{.EXEC_CMD}}

rm -f $CmdPath
`)