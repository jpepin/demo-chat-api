

db-up: 
	docker run --name message-mysql --network app-bridge --publish 3306:3306 -e MYSQL_ROOT_PASSWORD=my-secret-pw -e MYSQL_DATABASE=app -d mysql:latest
	echo "Allow db to initialize"
	sleep 10

db-clean:
	- docker stop message-mysql && docker rm message-mysql 

network: 
	docker network create app-bridge

network-clean:
	- docker network rm app-bridge

image:
	docker image build -t message-app .

run-app:
	docker container run -p 8080:8080 --network app-bridge message-app

run-message-api: db-clean db-up run-app

build-and-run: network image run-message-api

app-clean: 
	- docker stop message-app && docker rm message-app

teardown: app-clean db-clean network-clean
