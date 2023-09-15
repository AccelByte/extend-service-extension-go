# 5. Project Structure

This chapter offers an overview of the Guild Service's project structure. Understanding this structure will help you navigate the project and identify where to make changes as you add new features or fix bugs.

```bash
.
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── apidocs                                   # Generated OpenAPI spec from proto file
│   └── guildService.swagger.json
├── docs
├── go.mod
├── go.sum
├── main.go
├── pkg
│   ├── common
│   ├── pb                                     # Generated stub from proto file
│   ├── proto
│   │   ├── google                             # protobuf library 
│   │   ├── protoc-gen-openapiv2               # protobuf library
│   │   └── guildService.proto                 # your proto file, where you put API definition
│   ├── service
│   │   ├── guildService.go                    # the implementation for the api spec written in your proto fie 
│   │   ├── guildService_test.go
│   │   └── mocks                              # generated code by mockgen to help unit testing
│   │       ├── repo_mock.go
│   │       └── server_mock.go
│   └── storage
└── third_party
    ├── embed.go
    └── swagger-ui                             # directory containing swagger UI
```

The most important files and directories are:

- `Makefile`: This file contains scripts that automate tasks like building our service, running tests, and cleaning up.
- `Dockerfile`: The Dockerfile for our service. This is used by Docker to build a container image for our service.
- `docker-compose.yaml`: This file defines the services that make up our application, so they can be run together using Docker Compose.
- `go.mod` and `go.sum`: These files are used by Go's dependency management system. They list the dependencies of our project.
- `main.go`: This is the main entry point for our service. It initializes and starts our server and gateway.
- `pkg`: This directory contains the main code of our service.
- `pkg/common/gateway.go`: Contains the code for our gRPC-Gateway, which translates HTTP/JSON requests into gRPC and vice versa.
- `pkg/pb`: This directory contains the Go code that was generated from our .proto files by the protoc compiler.
- `pkg/proto`: This directory contains our .proto files, which define our gRPC service and messages.
- `pkg/service/guildService.go`: This directory contains the implementation of our gRPC service.
- `apidocs`: This is where the generated OpenAPI spec located.
- `third_party`: This directory contains third party libraries that are used by our service.

In the following chapters, we will discuss how to define and implement new services and messages in our `.proto` files, how to generate Go code from these `.proto` files, and how to implement these services in our server.