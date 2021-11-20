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
    # 
    if [ "${PM}" = "yum" ]; then
        yum update && yum install -y curl socat supervisor
    elif [ "${PM}" = "apt-get" ]; then
        apt-get update && apt-get install -y curl socat supervisor
    fi
    judge "安装脚本依赖"
	add_supervisor_config
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

add_supervisor_config() {
	#
	touch /root/.sdwan/work.sh
	cat > /root/.sdwan/work.sh <<-EOF
#!/bin/bash
SERVER_URL="{{.SERVER_URL}}"
NODE_NAME="{{.NODE_NAME}}"
NODE_TOKEN="{{.NODE_TOKEN}}"
NODE_MODE="host"

if [ -f "/root/.sdwan/share/sdos" ]; then
	mkdir -p /tmp/.sdwan/work/
	host=$(echo "$SERVER_URL" | awk -F "/" '{print $3}')
	exi=$(echo "$SERVER_URL" | grep 'https://')
	if [ -n "$exi" ]; then
		url="wss://${host}/ws"
	else
		url="ws://${host}/ws"
	fi
	chmod +x /root/.sdwan/share/sdos
	/root/.sdwan/share/sdos work --server-url="${url}?action=nodework&nodemode=${NODE_MODE}&nodename=${NODE_NAME}&nodetoken=${NODE_TOKEN}&hostname=${HOSTNAME}"
else
	echo "work file does not exist"
	sleep 3
	exit 1
fi
EOF
	chmod +x /root/.sdwan/work.sh
	#
	touch /etc/supervisor/conf.d/sdwan.conf
	cat > /etc/supervisor/conf.d/sdwan.conf <<-EOF
[program:sdwan]
directory=/root/.sdwan
command=/root/.sdwan/work.sh
numprocs=1
autostart=true
autorestart=true
startretries=3
user=root
redirect_stderr=true
stdout_logfile=/var/log/supervisor/%(program_name)s.log
EOF
	#
	if [ "${PM}" = "yum" ]; then
		systemctl start supervisord
    elif [ "${PM}" = "apt-get" ]; then
		systemctl start supervisor
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
    cat > ~/.bashrc_docker <<-EOF
dockeralias()
{
    local var=\$1
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
fi

echo "success" > /tmp/.sdwan_install
rm -f $CmdPath
`)
