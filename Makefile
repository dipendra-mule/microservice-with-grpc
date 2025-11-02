.PHONY: proto build docker-up docker-down test

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/user/user.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/order/order.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/product/product.proto

build:
	go build -o bin/user-service cmd/user-service/main.go
	go build -o bin/order-service cmd/order-service/main.go
	go build -o bin/product-service cmd/product-service/main.go
	go build -o bin/gateway cmd/gateway/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

test:
	go test ./...

clean:
	rm -rf bin/