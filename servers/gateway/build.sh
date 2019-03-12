GOOS=linux go build
docker build -t kjgoodwins/server-api .
go clean

cd ..
cd db
docker build -t kjgoodwins/mysql .

cd ..
cd messaging
docker build -t kjgoodwins/message-service .

cd ..
cd summary
GOOS=linux go build
docker build -t kjgoodwins/summary-service .
go clean


