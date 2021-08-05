package install

const dockerCompose = string(`version: '3'
services:
  sdos:
    container_name: "sdwan-manage"
    image: "kuaifan/sdwan:manage-0.0.1"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /usr/bin/docker:/usr/bin/docker
      - /root/.sdwan/config:/usr/sdwan/config
      - /etc/localtime:/etc/localtime:ro
    environment:
      SERVER_URL: "{{.SERVER_URL}}"
      NODE_NAME: "{{.NODE_NAME}}"
      NODE_MODE: "manage"
    restart: unless-stopped
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

HostIp="{{.HostIp}}"

source '/etc/os-release' > /dev/null

is_root() {
    if [ 0 == $UID ]; then
        echo -e "${OK} ${GreenBG} 当前用户是root用户，进入安装流程 ${Font}"
        sleep 3
    else
        echo -e "${Error} ${RedBG} 当前用户不是root用户，请切换到root用户后重新执行脚本 ${Font}"
        exit 1
    fi
}

judge() {
    if [[ 0 -eq $? ]]; then
        echo -e "${OK} ${GreenBG} $1 完成 ${Font}"
        sleep 1
    else
        echo -e "${Error} ${RedBG} $1 失败${Font}"
        exit 1
    fi
}

local_ip() {
  IP=$1
  if [[ $IP =~ ^192\.168\..* ]] || [[ $IP =~ ^172\..* ]] || [[ $IP =~ ^10\..* ]]; then
    IP=""
  fi
  [ "$(check_ip "$IP")" != "yes" ] && IP=""
  echo -e "$IP"
}

check_ip() {
    IP=$1
    VALID_CHECK=$(echo "$IP" | awk -F. '$1<=255&&$2<=255&&$3<=255&&$4<=255{print "yes"}')
    if echo "$IP" | grep -E "^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$" > /dev/null; then
        if [ ${VALID_CHECK:-no} == "yes" ]; then
            echo "yes"
        else
            echo "no"
        fi
    else
        echo "no"
    fi
}

check_system() {
    if [[ "${ID}" == "centos" && ${VERSION_ID} -ge 7 ]]; then
        echo -e "${OK} ${GreenBG} 当前系统为 Centos ${VERSION_ID} ${VERSION} ${Font}"
    elif [[ "${ID}" == "debian" && ${VERSION_ID} -ge 8 ]]; then
        echo -e "${OK} ${GreenBG} 当前系统为 Debian ${VERSION_ID} ${VERSION} ${Font}"
    elif [[ "${ID}" == "ubuntu" && $(echo "${VERSION_ID}" | cut -d '.' -f1) -ge 16 ]]; then
        echo -e "${OK} ${GreenBG} 当前系统为 Ubuntu ${VERSION_ID} ${UBUNTU_CODENAME} ${Font}"
    else
        echo -e "${Error} ${RedBG} 当前系统为 ${ID} ${VERSION_ID} 不在支持的系统列表内，安装中断 ${Font}"
        exit 1
    fi
}

check_docker() {
    echo -e "检查Docker......"
    docker --version &> /dev/null
    if [ $? -eq  0 ]; then
        echo -e "${OK} ${GreenBG} 检查到Docker已安装！ ${Font}"
    else
        echo -e "安装docker环境..."
        curl -sSL https://get.daocloud.io/docker | sh
        echo -e "${OK} ${GreenBG} Docker环境安装完成！ ${Font}"
    fi
    systemctl start docker
    judge "Docker 启动"
    #
    echo -e "检查Docker-compose......"
    docker-compose --version &> /dev/null
    if [ $? -eq  0 ]; then
        echo -e -e "${OK} ${GreenBG} 检查到Docker-compose已安装！ ${Font}"
    else
        echo -e "安装docker-compose..."
        sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
        echo -e "${OK} ${GreenBG} Docker-compose安装完成！ ${Font}"
        service docker restart
    fi
}

if [ "$1" = "init" ]; then
    check_system
    check_docker
    
    cd "$(dirname $0)"
    docker-compose up -d
fi
`)
