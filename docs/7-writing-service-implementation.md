# Chapter 7: Writing Service Implementations

Now that we have defined our service and generated the necessary Go files, 
the next step is to implement our service. 
This is where we define the actual logic of our gRPC methods. 
We'll be doing this in the `pkg/service/guildService.go` file.

Here's a brief outline of what this chapter will cover:

## 7.1 Setting Up the Guild Service

### 7.1 Setting Up the Guild Service
To set up our guild service, we'll first create an object that embeds the `UnimplementedGuildServiceServer` in our `guildService.go` file. This object will act as our service implementation.

```go
type GuildServiceServerImpl struct {
    pb.UnimplementedServiceServer
    // Other fields
}
```

This structure implements the `pb.UnimplementedServiceServer` interface, 
and holds relevant fields which will be used for our CloudSave setup later.

To implement the `CreateOrUpdateGuildProgress` function, your `GuildServiceServerImpl` would then 
need a method like this:

```go
func (s *GuildServiceServerImpl) CreateOrUpdateGuildProgress(
    ctx context.Context, req *pb.CreateOrUpdateGuildProgressRequest,
) (*pb.CreateOrUpdateGuildProgressResponse, error) {
	// Implementation goes here
}
```

And similarly for the GetGuildProgress function:

```go
func (s *GuildServiceServerImpl) GetGuildProgress(
    ctx context.Context, req *pb.GetGuildProgressRequest,
) (*pb.GetGuildProgressResponse, error) {
	// Implementation goes here
}
```

In these methods, you would include the logic to interact with CloudSave or 
any other dependencies in order to process the requests.