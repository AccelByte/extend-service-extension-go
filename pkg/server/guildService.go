// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/001extend/extend-custom-guild-service/pkg/pb"
	"github.com/AccelByte/accelbyte-go-sdk/cloudsave-sdk/pkg/cloudsaveclient/admin_game_record"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/cloudsave"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GuildServiceServer struct {
	pb.UnimplementedGuildServiceServer
	tokenRepo   repository.TokenRepository
	configRepo  repository.ConfigRepository
	refreshRepo repository.RefreshTokenRepository
}

func NewGuildServiceServer(
	tokenRepo *repository.TokenRepository, configRepo *repository.ConfigRepository,
	refreshRepo *repository.RefreshTokenRepository,
) (*GuildServiceServer, error) {
	return &GuildServiceServer{
		tokenRepo:   *tokenRepo,
		configRepo:  *configRepo,
		refreshRepo: *refreshRepo,
	}, nil
}

func (g GuildServiceServer) CreateOrUpdateGuildProgress(
	ctx context.Context, req *pb.CreateOrUpdateGuildProgressRequest,
) (*pb.CreateOrUpdateGuildProgressResponse, error) {
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

func (g GuildServiceServer) GetGuildProgress(
	ctx context.Context, req *pb.GetGuildProgressRequest,
) (*pb.GetGuildProgressResponse, error) {
	// Initialize the AccelByte CloudSave service
	adminGameRecordService := &cloudsave.AdminGameRecordService{
		Client:                 factory.NewCloudsaveClient(g.configRepo),
		TokenRepository:        g.tokenRepo,
		RefreshTokenRepository: g.refreshRepo,
	}

	// Get guild progress in CloudSave
	guildProgressKey := fmt.Sprintf("guildProgress_%s", req.GuildId)

	input := &admin_game_record.AdminGetGameRecordHandlerV1Params{
		Key:       guildProgressKey,
		Namespace: getNamespace(),
	}
	response, err := adminGameRecordService.AdminGetGameRecordHandlerV1Short(input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error updating guild progress: %v", err)
	}

	// Convert the response value to a JSON string
	valueJSON, err := json.Marshal(response.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error marshalling value into JSON: %v", err)
	}

	// Unmarshal the JSON string into a pb.GuildProgress
	var guildProgress pb.GuildProgress
	err = json.Unmarshal(valueJSON, &guildProgress)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error unmarshalling value into GuildProgress: %v", err)
	}

	return &pb.GetGuildProgressResponse{
		GuildProgress: &guildProgress,
	}, nil
}
