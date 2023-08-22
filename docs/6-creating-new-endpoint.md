# 6. Creating a New Endpoint

In this chapter, we will be adding new endpoints to our service. This involves three main steps:

1. Defining the service and its methods in our `.proto` file.
2. Generating Go code from the updated `.proto` file.
3. Implementing the new service methods in our server.

## 6.1 Defining the Service in the .proto File

gRPC services and messages are defined in `.proto` files. Our `.proto` file is located in `pkg/proto/guildService.proto`. Let's add new service methods to our `GuildService`:

```protobuf
service GuildService {
  rpc CreateOrUpdateGuildProgress (CreateOrUpdateGuildProgressRequest) returns (CreateOrUpdateGuildProgressResponse) {
    option (google.api.http) = {
      post: "/guild/v1/progress"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Update Guild progression"
      description: "Update Guild progression if not existed yet will create a new one"
    };
  }

  rpc GetGuildProgress (GetGuildProgressRequest) returns (GetGuildProgressResponse) {
    option (google.api.http) = {
      get: "/guild/v1/progress/{guild_id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "Get guild progression"
      description: "Get guild progression"
    };
  }
}

message CreateOrUpdateGuildProgressRequest {
  string guild_id = 1;
  GuildProgress guild_progress = 2;
}

message CreateOrUpdateGuildProgressResponse {
  GuildProgress guild_progress = 1;
}

message GetGuildProgressRequest {
  string guild_id = 1;
}

message GetGuildProgressResponse {
  GuildProgress guild_progress = 1;
}

message GuildProgress {
  string guild_id = 1;
  map<string, int32> objectives = 2;
}
```

In this case, we've added two service methods: `CreateOrUpdateGuildProgress` and `GetGuildProgress`.

In the CreateOrUpdateGuildProgress method, we use the `option (google.api.http)` annotation 
to specify the HTTP method and path for this service method. 
The post: `"/guild/v1/progress"` means that this service method will be exposed as a 
POST HTTP method at the path `"/guild/v1/progress"`.

We use body: `"*"` to indicate that the entire request message will be used as the 
HTTP request body. Alternatively, you could specify a particular field of the 
request message to be used as the HTTP request body.

The `option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation)` annotation 
is used for additional metadata about the operation that can be used by tools like Swagger.

Similarly, in the `GetGuildProgress` method, we specify a GET HTTP method and 
the path includes a variable part `{guild_id}` which will be substituted with the actual 
`guild_id` in the HTTP request.

After defining the service and methods in the `.proto` file, we run the protoc compiler 
to generate the corresponding Go code.

## 6.2 Generating Go Code

After updating our .proto file, we need to generate Go code from it.
The protobuf compiler `protoc` is used to generate Go code from our .proto file. 
However, in our setup, we've simplified this with a `Makefile`.


```bash
make proto
```
