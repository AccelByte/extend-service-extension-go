# Chapter 8: Integrating with AccelByte's CloudSave

In this chapter, we'll learn how to integrate the AccelByte's CloudSave feature into our GuildService.

## 8.1. Understanding CloudSave

AccelByte's CloudSave is a cloud-based service that enables you to save and retrieve game data in 
a structured manner. It allows for easy and quick synchronization of player data across different 
devices. This can be especially useful in multiplayer games where players' data needs to be synced 
in real-time. Please refer to our docs portal for more details

## 8.2. Setting up CloudSave

The first step to using CloudSave is setting it up. 
In the context of our GuildService, this involves adding the CloudSave client to our server struct 
and initializing it during server startup.

```go
type GuildServiceServerImpl struct {
	pb.UnimplementedGuildServiceServer
	tokenRepo   repository.TokenRepository
	configRepo  repository.ConfigRepository
	refreshRepo repository.RefreshTokenRepository
}
```

During server startup, you would initialize the requirement of CloudSave client like so:

```go
// Preparing the IAM authorization
var tokenRepo repository.TokenRepository = sdkAuth.DefaultTokenRepositoryImpl()
var configRepo repository.ConfigRepository = sdkAuth.DefaultConfigRepositoryImpl()
var refreshRepo repository.RefreshTokenRepository = sdkAuth.DefaultRefreshTokenImpl()

// Configure IAM authorization
oauthService := iam.OAuth20Service{
    Client:                 factory.NewIamClient(configRepo),
    TokenRepository:        tokenRepo,
    RefreshTokenRepository: refreshRepo,
}

clientId := configRepo.GetClientId()
clientSecret := configRepo.GetClientSecret()
err := oauthService.LoginClient(&clientId, &clientSecret)
if err != nil {
    logrus.Fatalf("Error unable to login using clientId and clientSecret: %v", err)
}

```

## 8.3. Using CloudSave in GuildService

Let's go over an example of how we use CloudSave within our GuildService.

When updating the guild progress, after performing any necessary validations and computations, 
you would save the updated progress to CloudSave like so:


```go
func (s *GuildServiceServerImpl) CreateOrUpdateGuildProgress(
    ctx context.Context, req *pb.CreateOrUpdateGuildProgressRequest,
) (*pb.CreateOrUpdateGuildProgressResponse, error) {
    // Other implementation, like your computation, validation, etc...
	
    // Initialize the AccelByte CloudSave service
    adminGameRecordService := &cloudsave.AdminGameRecordService{
        Client:                 factory.NewCloudsaveClient(g.configRepo), 
        TokenRepository:        g.tokenRepo,
        RefreshTokenRepository: g.refreshRepo,
    }

    // Create or update guild progress in CloudSave
    // This assumes we're storing guild progress as a JSON object
    guildProgressKey := fmt.Sprintf("guildProgress_%s", req.GuildId)
    guildProgressValue := req.GuildProgress

    input := &admin_game_record.AdminPostGameRecordHandlerV1Params{
        Body:      guildProgressValue,
        Key:       guildProgressKey,
        Namespace: getNamespace(),
    }
    _, err := adminGameRecordService.AdminPostGameRecordHandlerV1Short(input)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "Error updating guild progress: %v", err)
    }

    // Return the updated guild progress
    return &pb.CreateOrUpdateGuildProgressResponse{GuildProgress: req.GuildProgress}, nil
}
```

That's it! You've now integrated AccelByte's CloudSave into your GuildService. 
You can now use CloudSave to save and retrieve guild progress, along with any other 
data you might need to store.