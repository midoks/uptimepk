#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


# curl -fsSL  https://raw.githubusercontent.com/midoks/uptimepk/master/scripts/install_dev.sh | sh

# Linux 手动安装
# wget https://go.dev/dl/go1.19.1.linux-amd64.tar.gz
# sudo tar -C /usr/local -xzf go1.19.1.linux-amd64.tar.gz
# sudo ln -s /usr/local/go/bin/* /usr/bin/

# systemctl status uptimepk

# 手动编译
# go build main.go -o uptimepk && uptimepk web 

# Debug Now
export PATH=/usr/local/go:$PATH:/root/go/bin
export GOPATH=/root/go


TAGRT_DIR=/usr/local/uptimepk_dev
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

export GIT_COMMIT=$(git rev-parse HEAD)
export BUILD_TIME=$(date -u '+%Y-%m-%d %I:%M:%S %Z')

go install github.com/midoks/zzz@latest

if [ ! -d $TAGRT_DIR/uptimepk ]; then
	git clone https://github.com/midoks/uptimepk
	cd $TAGRT_DIR/uptimepk
else
	cd $TAGRT_DIR/uptimepk
	git pull https://github.com/midoks/uptimepk
fi

go mod tidy
go mod vendor

# cd /usr/local/uptimepk_dev/uptimepk && go build -o uptimepk main.go 
cd $TAGRT_DIR/uptimepk && go build -o uptimepk main.go 


cd $TAGRT_DIR/uptimepk/scripts

sh make.sh


systemctl daemon-reload


# rm -rf /usr/local/uptimepk_dev/uptimepk/custom
# rm -rf /usr/local/uptimepk_dev/uptimepk/data

service uptimepk restart

cd $TAGRT_DIR/uptimepk && ./uptimepk -v

if [ ! -d /usr/local/go ];then
	wget https://golang.google.cn/dl/go1.26.2.linux-amd64.tar.gz
	tar -xvf go1.26.2.linux-amd64.tar.gz
	mv go /usr/local/
fi


if [ ! -f /root/go/bin/zzz ];then
	go install github.com/midoks/zzz@latest
fi

service uptimepk stop

