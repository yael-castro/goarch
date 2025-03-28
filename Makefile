.PHONY: http relay proto up down

http:
	sh scripts/build.bash http

relay:
	sh scripts/build.bash relay

proto:
	bash scripts/proto.bash

up:
	docker-compose up --build

down:
	docker-compose down -v --rmi local --remove-orphans
