docker run -d \
-p 3306:3306 \
--name mysqlserver \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=website \
kjgoodwins/mysql