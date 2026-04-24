[Unit]
Description=UptimePK Server
After=network.service
After=syslog.target

[Service]
User=root
Group=root
Type=simple
WorkingDirectory={APP_PATH}
ExecStart=uptimepk web
ExecReload=/bin/kill -USR2 $MAINPID
PermissionsStartOnly=true
LimitNOFILE=65535
Restart=on-failure
RestartSec=10
RestartPreventExitStatus=1
PrivateTmp=false


[Install]
WantedBy=multi-user.target