run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

docker:
	docker build -t trustify-badge-backend:latest .
