

db-up: 
	docker run --name message-mysql  --publish 3306:3306 -e MYSQL_ROOT_PASSWORD=my-secret-pw -e MYSQL_DATABASE=app -d mysql:latest

db-clean:
	docker stop message-mysql && docker rm message-mysql 

