#! /bin/bash

read -p 'APPLICATION TO RUN [local/web/was]: ' app
if [ $app == 'was' ]
then
	read -p 'DB Name: ' DBNAME
	read -p 'DB User Name: ' DBUSERNAME
	read -p 'DB Password: ' DBPASSWORD
fi


appPath=${PWD}
systemPath=/etc/systemd/system/landingProject.service


sudo printf "[Unit]\nDescription=SRE Landing Project Application Daemon Service\nAfter=network.target\n[Service]\nType=Simple\nExecStart=$appPath/$app/$app\nRestart=always" >> $systemPath

sudo systemctl start landingProject.service
