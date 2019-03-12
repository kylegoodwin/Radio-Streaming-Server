sh build.sh
docker push kjgoodwins/server-api
docker push kjgoodwins/mysql
ssh root@138.68.18.8 << EOF
docker rm -f kjgoodwins

docker rm -f api

docker run -d -p 443:443 --network site -v /etc/letsencrypt:/etc/letsencrypt \
-e TLSKEY=/etc/letsencrypt/live/info-api.kylegoodwin.net/privkey.pem \
-e TLSCERT=/etc/letsencrypt/live/info-api.kylegoodwin.net/fullchain.pem \
-e MYSQL_ROOT_PASSWORD=password \
-e SESSIONKEY=sessionkey \
-e REDISADDR=redis:6379 \
-e DBADDR=mysqlserver:3306 \
--name api \
kjgoodwins/server-api

EOF
