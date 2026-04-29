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
DATA_DIR="${INSTALL_DIR}/custom/data"
CONF_FILE="${INSTALL_DIR}/custom/conf"
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
        yum install -y wget tar
    elif [ "$OS" = "debian" ]; then
        apt-get install -y wget tar
    else
        echo -e "${YELLOW}跳过依赖安装，请确保已安装 wget 和 tar${NC}"
    fi
}

# 创建目录结构
create_dirs() {
    echo -e "${YELLOW}创建目录结构...${NC}"
    mkdir -p "$INSTALL_DIR"
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

# 创建系统服务
create_service() {
    echo -e "${YELLOW}创建系统服务...${NC}"
    
    cd $INSTALL_DIR && ./uptimepk install

    echo -e "${GREEN}服务创建完成${NC}"
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
    echo -e "  http://$(hostname -I | awk '{print $1}'):9191"
    echo -e "  管理后台: http://$(hostname -I | awk '{print $1}'):9191/uptimepk"
    echo ""
    echo -e "${YELLOW}服务管理:${NC}"
    echo -e "  启动: systemctl start ${SERVICE_NAME}"
    echo -e "  停止: systemctl stop ${SERVICE_NAME}"
    echo -e "  重启: systemctl restart ${SERVICE_NAME}"
    echo -e "  状态: systemctl status ${SERVICE_NAME}"
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
    
    # 创建服务
    create_service
    
    # 启动服务
    systemctl start uptimepk
    
    # 显示信息
    show_info
}

# 执行主流程
main "$@"
