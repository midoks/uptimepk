#!/bin/bash
# ============================================
# uptimepk 一键安装脚本
# 支持系统: CentOS 7+, Ubuntu 16.04+, Debian 9+
# 支持架构: x86_64, i386, arm64
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 版本信息
VERSION="v1.0.6"
REPO="midoks/uptimepk"
INSTALL_DIR="/opt/uptimepk"
DATA_DIR="/var/lib/uptimepk"
CONF_FILE="/etc/uptimepk.conf"
SERVICE_NAME="uptimepk"

# 检测架构
detect_arch() {
    case "$(uname -m)" in
        x86_64) ARCH="amd64" ;;
        i386|i686) ARCH="386" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) 
            echo -e "${RED}不支持的架构: $(uname -m)${NC}"
            exit 1
            ;;
    esac
    echo -e "${GREEN}检测到架构: ${ARCH}${NC}"
}

# 检测操作系统
detect_os() {
    if [ -f /etc/centos-release ] || [ -f /etc/redhat-release ]; then
        OS="centos"
    elif [ -f /etc/debian_version ] || [ -f /etc/lsb-release ]; then
        OS="debian"
    else
        echo -e "${YELLOW}无法检测操作系统，将尝试通用安装方式${NC}"
        OS="unknown"
    fi
    echo -e "${GREEN}检测到操作系统: ${OS}${NC}"
}

# 安装依赖
install_deps() {
    echo -e "${YELLOW}正在安装依赖...${NC}"
    if [ "$OS" = "centos" ]; then
        yum update -y && yum install -y wget tar
    elif [ "$OS" = "debian" ]; then
        apt-get update -y && apt-get install -y wget tar
    else
        echo -e "${YELLOW}跳过依赖安装，请确保已安装 wget 和 tar${NC}"
    fi
}

# 创建目录结构
create_dirs() {
    echo -e "${YELLOW}创建目录结构...${NC}"
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p /etc/uptimepk
}

# 下载并解压
download_and_extract() {
    echo -e "${YELLOW}正在下载 uptimepk ${VERSION}...${NC}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/uptimepk_${VERSION}_linux_${ARCH}.tar.gz"
    TMP_FILE=$(mktemp)
    
    if ! wget -q -O "$TMP_FILE" "$URL"; then
        echo -e "${RED}下载失败，请检查网络连接${NC}"
        rm -f "$TMP_FILE"
        exit 1
    fi
    
    echo -e "${YELLOW}正在解压...${NC}"
    tar -xzf "$TMP_FILE" -C "$INSTALL_DIR"
    rm -f "$TMP_FILE"
    
    chmod +x "$INSTALL_DIR/uptimepk"
    echo -e "${GREEN}解压完成${NC}"
}

# 创建配置文件
create_config() {
    echo -e "${YELLOW}创建配置文件...${NC}"
    
    # 询问用户配置
    read -p "请输入监听端口 (默认: 9191): " PORT
    PORT=${PORT:-9191}
    
    read -p "请输入管理员用户名 (默认: admin): " ADMIN_USER
    ADMIN_USER=${ADMIN_USER:-admin}
    
    read -p "请输入管理员密码 (默认: admin123): " ADMIN_PASS
    ADMIN_PASS=${ADMIN_PASS:-admin123}
    
    cat > "$CONF_FILE" << EOF
# uptimepk 配置文件

# 服务配置
[server]
port = ${PORT}
admin_path = "admin"

# 数据库配置
[database]
type = "sqlite"
path = "${DATA_DIR}/uptimepk.db"

# 管理员配置
[admin]
username = "${ADMIN_USER}"
password = "${ADMIN_PASS}"

# 日志配置
[log]
level = "info"
path = "${DATA_DIR}/logs"
EOF
    
    echo -e "${GREEN}配置文件创建完成${NC}"
}

# 创建系统服务
create_service() {
    echo -e "${YELLOW}创建系统服务...${NC}"
    
    if [ "$OS" = "centos" ]; then
        # CentOS 7+ systemd
        cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=uptimepk - Website monitoring service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/uptimepk -c ${CONF_FILE}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable ${SERVICE_NAME}
    elif [ "$OS" = "debian" ]; then
        # Debian/Ubuntu systemd
        cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=uptimepk - Website monitoring service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/uptimepk -c ${CONF_FILE}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable ${SERVICE_NAME}
    else
        # 通用 init.d 脚本
        cat > /etc/init.d/${SERVICE_NAME} << 'EOF'
#!/bin/bash
# chkconfig: 2345 90 10
# description: uptimepk - Website monitoring service

### BEGIN INIT INFO
# Provides:          uptimepk
# Required-Start:    $network
# Required-Stop:     $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Start/stop uptimepk monitoring service
### END INIT INFO

INSTALL_DIR="/opt/uptimepk"
CONF_FILE="/etc/uptimepk.conf"
PID_FILE="/var/run/uptimepk.pid"

case "$1" in
    start)
        echo "Starting uptimepk..."
        cd $INSTALL_DIR && nohup ./uptimepk -c $CONF_FILE > /var/log/uptimepk.log 2>&1 &
        echo $! > $PID_FILE
        ;;
    stop)
        echo "Stopping uptimepk..."
        kill -TERM $(cat $PID_FILE) 2>/dev/null || true
        rm -f $PID_FILE
        ;;
    restart)
        $0 stop
        sleep 2
        $0 start
        ;;
    status)
        if [ -f $PID_FILE ] && kill -0 $(cat $PID_FILE) 2>/dev/null; then
            echo "uptimepk is running"
            exit 0
        else
            echo "uptimepk is not running"
            exit 1
        fi
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
exit 0
EOF
        chmod +x /etc/init.d/${SERVICE_NAME}
        if command -v chkconfig &>/dev/null; then
            chkconfig --add ${SERVICE_NAME}
        fi
    fi
    
    echo -e "${GREEN}服务创建完成${NC}"
}

# 启动服务
start_service() {
    echo -e "${YELLOW}启动服务...${NC}"
    if [ "$OS" = "centos" ] || [ "$OS" = "debian" ]; then
        systemctl start ${SERVICE_NAME}
        sleep 3
        if systemctl is-active --quiet ${SERVICE_NAME}; then
            echo -e "${GREEN}服务启动成功${NC}"
        else
            echo -e "${RED}服务启动失败，请检查日志${NC}"
            exit 1
        fi
    else
        /etc/init.d/${SERVICE_NAME} start
        sleep 3
        if /etc/init.d/${SERVICE_NAME} status; then
            echo -e "${GREEN}服务启动成功${NC}"
        else
            echo -e "${RED}服务启动失败，请检查日志${NC}"
            exit 1
        fi
    fi
}

# 显示安装信息
show_info() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}     uptimepk 安装完成！${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${YELLOW}服务信息:${NC}"
    echo -e "  服务名称: ${SERVICE_NAME}"
    echo -e "  安装目录: ${INSTALL_DIR}"
    echo -e "  数据目录: ${DATA_DIR}"
    echo -e "  配置文件: ${CONF_FILE}"
    echo ""
    echo -e "${YELLOW}访问地址:${NC}"
    echo -e "  http://$(hostname -I | awk '{print $1}'):${PORT}"
    echo -e "  管理后台: http://$(hostname -I | awk '{print $1}'):${PORT}/admin"
    echo ""
    echo -e "${YELLOW}登录信息:${NC}"
    echo -e "  用户名: ${ADMIN_USER}"
    echo -e "  密码: ${ADMIN_PASS}"
    echo ""
    echo -e "${YELLOW}服务管理:${NC}"
    if [ "$OS" = "centos" ] || [ "$OS" = "debian" ]; then
        echo -e "  启动: systemctl start ${SERVICE_NAME}"
        echo -e "  停止: systemctl stop ${SERVICE_NAME}"
        echo -e "  重启: systemctl restart ${SERVICE_NAME}"
        echo -e "  状态: systemctl status ${SERVICE_NAME}"
    else
        echo -e "  启动: /etc/init.d/${SERVICE_NAME} start"
        echo -e "  停止: /etc/init.d/${SERVICE_NAME} stop"
        echo -e "  重启: /etc/init.d/${SERVICE_NAME} restart"
        echo -e "  状态: /etc/init.d/${SERVICE_NAME} status"
    fi
    echo ""
    echo -e "${GREEN}========================================${NC}"
}

# 主安装流程
main() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}    uptimepk 一键安装脚本${NC}"
    echo -e "${GREEN}        Version: ${VERSION}${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    # 检查是否为 root 用户
    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi
    
    # 检测架构和操作系统
    detect_arch
    detect_os
    
    # 安装依赖
    install_deps
    
    # 创建目录
    create_dirs
    
    # 下载解压
    download_and_extract
    
    # 创建配置
    create_config
    
    # 创建服务
    create_service
    
    # 启动服务
    start_service
    
    # 显示信息
    show_info
}

# 执行主流程
main "$@"
