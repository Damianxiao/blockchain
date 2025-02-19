build:
	go build -o ./app/blockchain ./cmd

run: build
	./app/blockchain

test:
	go test  ./...

proto:
	rm -rf ./idl/pb/core.pb.go && protoc ./idl/*.proto  --go_out=./idl/pb