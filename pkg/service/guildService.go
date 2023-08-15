// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	pb "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/storage"
	"fmt"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GuildServiceServerImpl struct {
	pb.UnimplementedGuildServiceServer
	tokenRepo   repository.TokenRepository
	configRepo  repository.ConfigRepository
	refreshRepo repository.RefreshTokenRepository
	storage     storage.Storage
}

func NewGuildServiceServer(
	tokenRepo repository.TokenRepository,
	configRepo repository.ConfigRepository,
	refreshRepo repository.RefreshTokenRepository,
	storage storage.Storage,
) *GuildServiceServerImpl {
	return &GuildServiceServerImpl{
		tokenRepo:   tokenRepo,
		configRepo:  configRepo,
		refreshRepo: refreshRepo,
		storage:     storage,
	}
}

func (g GuildServiceServerImpl) CreateOrUpdateGuildProgress(
	ctx context.Context, req *pb.CreateOrUpdateGuildProgressRequest,
) (*pb.CreateOrUpdateGuildProgressResponse, error) {
	// Create or update guild progress in CloudSave
	// This assumes we're storing guild progress as a JSON object
	guildProgressKey := fmt.Sprintf("guildProgress_%s", req.GuildId)
	guildProgressValue := req.GuildProgress
	err := g.storage.SaveGuildProgress(guildProgressKey, guildProgressValue)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error updating guild progress: %v", err)
	}

	// Return the updated guild progress
	return &pb.CreateOrUpdateGuildProgressResponse{GuildProgress: req.GuildProgress}, nil
}

func (g GuildServiceServerImpl) GetGuildProgress(
	ctx context.Context, req *pb.GetGuildProgressRequest,
) (*pb.GetGuildProgressResponse, error) {
	// Get guild progress in CloudSave
	guildProgressKey := fmt.Sprintf("guildProgress_%s", req.GuildId)

	guildProgress, err := g.storage.GetGuildProgress(guildProgressKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error getting guild progress: %v", err)
	}

	return &pb.GetGuildProgressResponse{
		GuildProgress: guildProgress,
	}, nil
}
