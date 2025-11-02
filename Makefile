.PHONY: run test

run:
	go run cmd/main.go

test:
	go test -count=1 ./... -v