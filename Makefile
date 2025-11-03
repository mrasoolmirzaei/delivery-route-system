.PHONY: run test

run:
	go run cmd/main.go

test:
	go test -count=1 ./... -v

build:
	docker build -t delivery-route-system .

run-docker:
	docker run -p 8000:8000 delivery-route-system