cd api-server
docker build -t kjgoodwins/audio-api .
cd ..

cd socket-server
docker build -t kjgoodwins/audio-socket .
cd ..



docker push kjgoodwins/audio-api
docker push kjgoodwins/audio-socket

export TLSCERT=/etc/letsencrypt/live/audio-api.kjgoodwin.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/audio-api.kjgoodwin.me/privkey.pem

ssh -tt ec2-user@18.211.146.1<< EOF
docker rm -f audio-api
docker rm -f audio-socket
docker pull kjgoodwins/audio-api
docker pull kjgoodwins/audio-socket
docker run -d \
--network site \
--name audio-socket \
-p 3001:3001 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=/etc/letsencrypt/live/audio-api.kjgoodwin.me/fullchain.pem \
-e TLSKEY=/etc/letsencrypt/live/audio-api.kjgoodwin.me/privkey.pem \
-e MDPORT=mongodb:27017 \
kjgoodwins/audio-socket
docker run -d \
--name audio-api \
--network site \
-e MDPORT=mongodb:27017 \
kjgoodwins/audio-api
exit
EOF