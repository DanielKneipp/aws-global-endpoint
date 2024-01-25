.PHONY: build

build: http-latency-test.go
	go build -o ./http-latency-test .
	GOOS=linux GOARCH=amd64 go build -o ./http-latency-test-linux-x64 .
