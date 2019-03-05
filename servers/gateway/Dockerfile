#Use this command when running the docker image from EC2 instance
#sudo docker run -p 443:443 -v /etc/letsencrypt:/etc/letsencrypt:ro -e TLSCERT=$TLSCERT -e TLSKEY=$TLSKEY --name zanewebb zanewebb/zanewebbuw
#TLSCERT must be an env like: export TLSCERT=/etc/letsencrypt/live/api.zanewebb.me/fullchain.pem
#TLSKEY must be an env like: export TLSKEY=/etc/letsencrypt/live/api.zanewebb.me/privkey.pem
FROM alpine
RUN apk add --no-cache ca-certificates
COPY gateway /gateway
EXPOSE 443
ENTRYPOINT [ "/gateway" ]