sh build.sh
docker push kjgoodwins/server-api
docker push kjgoodwins/mysql
docker push kjgoodwins/message-service-new

ssh -tt ec2-user@18.211.146.1 << EOF
docker rm -f api
docker rm -f message1
docker rm -f summary1

docker pull kjgoodwins/server-api
docker pull kjgoodwins/mysql
docker pull kjgoodwins/message-service-new

docker network create site

docker run -d --name redis --network site redis

docker run -d \
--network site \
--name mysqlserver \
-e MYSQL_ROOT_PASSWORD=password \
-e MYSQL_DATABASE=website \
kjgoodwins/mysql

docker run -d --hostname rabbit --name rabbit \
--network site -p 15672:15672 rabbitmq:3-management

docker run -d \
--network site \
--name mongodb \
mongo

docker run -d \
--network site \
--name message1 \
-e NAME=message1 \
-e PORT=5001 \
-e MONGOPORT=mongodb:27017 \
kjgoodwins/message-service-new

docker run -d -p 443:443 --network site -v /etc/letsencrypt:/etc/letsencrypt \
-e TLSKEY=/etc/letsencrypt/live/audio-api.kjgoodwin.me/privkey.pem \
-e TLSCERT=/etc/letsencrypt/live/audio-api.kjgoodwin.me/fullchain.pem \
-e MYSQL_ROOT_PASSWORD=password \
-e SESSIONKEY=sessionkey \
-e REDISADDR=redis:6379 \
-e DBADDR=mysqlserver:3306 \
-e MESSAGESADDRS=http://message1:5001 \
-e SUMMARYADDRS=http://summary1:5003 \
--name api \
kjgoodwins/server-api

exit
EOF
