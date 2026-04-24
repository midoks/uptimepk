#!/bin/bash

_os=`uname`
_path=`pwd`
_dir=`dirname $_path`

sed "s:{APP_PATH}:${_dir}:g" $_dir/scripts/init.d/uptimepk.tpl > $_dir/scripts/init.d/uptimepk
chmod +x $_dir/scripts/init.d/uptimepk

sed "s:{APP_PATH}:${_dir}:g" $_dir/scripts/init.d/uptimepk.service.tpl > $_dir/scripts/init.d/uptimepk.service

if [ -d /etc/init.d ];then
	cp $_dir/scripts/init.d/uptimepk /etc/init.d/uptimepk
	chmod +x /etc/init.d/uptimepk
fi

echo `dirname $_path`