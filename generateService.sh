#! /bin/bash

read -p 'APPLICATION TO RUN [local/web/was]: ' app
appPath=${PWD}

service="[Unit]
Description=SRE Landing Project Application Daemon Service
After=network.target

[Service]
Type=Simple
ExecStart=sudo $appPath/$app/$app
Restart=always
"

sudo echo $service >> /etc/systemd/system/landingProject.service

sudo systemctl start landingProject.service
