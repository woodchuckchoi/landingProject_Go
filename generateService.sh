#! /bin/bash

read -p 'APPLICATION TO RUN [local/web/was]: ' app

appPath=${PWD}
systemPath=/etc/systemd/system/landingProject.service


sudo echo "[Unit]" >> $systemPath
sudo echo "Description=SRE Landing Project Application Daemon Service" >> $systemPath
sudo echo "After=network.target" >> $systemPath
sudo echo "[Service]" >> $systemPath
sudo echo "Type=Simple" >> $systemPath
sudo echo "ExecStart=sudo $appPath/$app/$app" >> $systemPath
sudo echo "Restart=always" >> $systemPath

#sudo systemctl start landingProject.service
