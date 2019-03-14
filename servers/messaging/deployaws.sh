docker build -t kjgoodwins/message-service-new .

docker push kjgoodwins/message-service-new

ssh -tt ec2-user@18.211.146.1 << EOF
docker rm -f message1

docker pull kjgoodwins/message-service-new

docker run -d \
--network site \
--name message1 \
-e NAME=message1 \
-e PORT=5001 \
-e MONGOPORT=mongodb:27017 \
kjgoodwins/message-service-new

exit
EOF
