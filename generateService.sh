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

if [ -f $systemPath ] 
then
	rm $systemPath
fi

printf "[Unit]\nDescription=SRE Landing Project Application Daemon Service\nAfter=network.target\n[Service]\nType=simple\nExecStart=$appPath/$app/$app $DBNAME $DBUSERNAME $DBPASSWORD\nRestart=always" >> $systemPath
systemctl daemon-reload
systemctl start landingProject.service
