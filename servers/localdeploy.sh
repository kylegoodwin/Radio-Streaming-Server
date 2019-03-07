export TLSCERT=FIGUREOUTTHISPATH
export TLSKEY=FIGUREOUTTHISPATH
export DSN="root:PASSWORDGOESHERE@tcp(mysqlServer:3306)/db"
export SESSIONKEY=sessionkey
export MYSQL_ROOT_PASSWORD=MAKEAPASSWORD
export MESSAGESADDR=http://messaging:80

sudo docker run -d --name redisServer --network gatewayNetwork redis
sudo docker run -d --name mysqlServer --network gatewayNetwork -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD -e MYSQL_DATABASE=db zanewebb/zanemysql
sudo docker run -d --name --network gatewayNetwork mongo mongo
sudo docker run -d --name messaging --network gatewayNetwork zanewebb/messaging
sudo docker run -d -p 443:443 -v /etc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY -e DSN=$DSN -e SESSIONKEY=$SESSIONKEY -e SUMMARYADDR=$SUMMARYADDR -e MESSAGESADDR=$MESSAGESADDR --name gateway --network gatewayNetwork zanewebb/zanewebbuw

