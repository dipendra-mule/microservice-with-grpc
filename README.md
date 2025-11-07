# Microservice with gRPC

A boilerplate Go project demonstrating a simple microservice architecture using gRPC for communication. This repository is ideal for learning or as a starting point for new Go gRPC-based backends.

## Features

- Service-to-service communication with gRPC
- Basic project scaffolding for easy extension
- Separation of business logic and transport
- Clean and idiomatic Go code

## Project Structure

```
microservice-with-grpc/
├── api/                # Protocol buffer definitions (*.proto) and generated Go code
├── cmd/                # Service entrypoints (main.go per service)
├── internal/           # Application/business logic
│   └── yourservice/    # Example service implementation
├── pb/                 # Compiled Protobuf packages
├── scripts/            # Helper scripts (build, run, etc.)
├── Dockerfile          # Dockerfile to containerize the service
├── go.mod / go.sum     # Go module files
└── README.md           # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.19 or newer
- `protoc` (Protocol Buffers compiler) and Go plugins (for gRPC)
- Docker (optional, for containerization)

### Clone the repository

```bash
git clone <repository-url>
cd microservice-with-grpc
```

### Generate Protobuf files

Before building the service, generate the Go code from your `.proto` files:

```bash
protoc --go_out=pb --go-grpc_out=pb api/your_service.proto
```

_(Repeat for each proto file in `api/` directory.)_

### Run the service

```bash
go run ./cmd/yourservice/main.go
```

Or build the binary:

```bash
go build -o yourservice ./cmd/yourservice
./yourservice
```

### Docker

To build and run the service in Docker:

```bash
docker build -t microservice-grpc .
docker run -p 50051:50051 microservice-grpc
```

## Extending the Project

- Add new `.proto` files in the `api/` directory for each service
- Implement corresponding server logic in `internal/`
- Register service handlers in the relevant `cmd/yourservice/main.go` file

## License

MIT License

