# 4. Installation and Setup

This chapter guides you through setting up your development environment. 
This involves installing required software, cloning the project repository, and 
setting up the project.

## 4.1. Software Installation

To get started, make sure you have the following software installed on your system:

1. [Go](https://golang.org/dl/): Please refer to the official installation guide.
2. [Protocol Buffers (protobuf)](https://developers.google.com/protocol-buffers): Protobuf is a language-agnostic binary data format developed by Google. It's used for defining the contract for our service.
3. [gRPC Gateway](https://github.com/grpc-ecosystem/grpc-gateway): gRPC Gateway generates a reverse-proxy server which translates a RESTful JSON API into gRPC.
4. [Docker](https://www.docker.com/products/docker-desktop): Docker is a platform for developers to develop, deploy, and run applications with containers.
5. [Docker Compose](https://docs.docker.com/compose/install/): Docker Compose is a tool for defining and running multi-container Docker applications.
6. [Make](https://www.gnu.org/software/make/): Make is a build automation tool.

### 4.2. Cloning and Running the dependency repo

This repository contains all dependencies that needed to be run for our service. 
So, ensure these dependencies is up and running before you run the guild service

```bash
$ git clone https://github.com/AccelByte/grpc-plugin-dependencies.git
```

```bash
$ cd grpc-plugin-dependencies
$ docker-compose up
```
