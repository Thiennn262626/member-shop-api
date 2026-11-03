build:
	go build -o bin/web-api cmd/main.go

run:
	go run cmd/main.go

format:
	go fmt web-api/...

build-docker:
	docker build . -t web-api

run-docker:
	docker run -itd --name web-api --restart always -p 8081:8081 web-api

exec-docker:
	docker exec -it web-api sh