package service

import (
	"context"
	"errors"
	pb "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/service/mocks"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

//go:generate mockgen -destination ./mocks/server_mock.go -package mocks extend-custom-guild-service/pkg/pb GuildServiceServer
//go:generate mockgen -destination ./mocks/repo_mock.go -package mocks github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository TokenRepository,ConfigRepository,RefreshTokenRepository

type cloudsaveStorageMock struct {
	mock.Mock
}

func (c *cloudsaveStorageMock) GetGuildProgress(key string) (*pb.GuildProgress, error) {
	args := c.Called(key)
	return args.Get(0).(*pb.GuildProgress), args.Error(1)
}

func (c *cloudsaveStorageMock) SaveGuildProgress(key string, value *pb.GuildProgress) error {
	args := c.Called(key, value)
	return args.Error(0)
}

func TestGuildServiceServerImpl_CreateOrUpdateGuildProgress(t *testing.T) {
	tests := []struct {
		name            string
		req             *pb.CreateOrUpdateGuildProgressRequest
		wantErr         bool
		expectedErr     error
		expectedGuildId string
	}{
		{
			name: "successful save",
			req: &pb.CreateOrUpdateGuildProgressRequest{
				GuildId: "testId",
				GuildProgress: &pb.GuildProgress{
					GuildId:    "testId",
					Objectives: map[string]int32{"testGoal": 1},
				},
			},
			wantErr:         false,
			expectedGuildId: "testId",
		},
		{
			name: "failed save",
			req: &pb.CreateOrUpdateGuildProgressRequest{
				GuildId: "testId",
				GuildProgress: &pb.GuildProgress{
					GuildId:    "testId",
					Objectives: map[string]int32{"testGoal": 1},
				},
			},
			wantErr:     true,
			expectedErr: errors.New("failed to save"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokenRepo := mocks.NewMockTokenRepository(ctrl)
			refreshRepo := mocks.NewMockRefreshTokenRepository(ctrl)
			configRepo := mocks.NewMockConfigRepository(ctrl)
			storage := new(cloudsaveStorageMock)
			service := NewGuildServiceServer(tokenRepo, configRepo, refreshRepo, storage)

			guildProgressKey := fmt.Sprintf("guildProgress_%s", tt.req.GuildId)
			storage.On("SaveGuildProgress", guildProgressKey, tt.req.GuildProgress).Return(tt.expectedErr)

			// when
			res, err := service.CreateOrUpdateGuildProgress(context.Background(), tt.req)

			// then
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedGuildId, res.GuildProgress.GuildId)
			}
			storage.AssertExpectations(t)
		})
	}
}

func TestGuildServiceServerImpl_GetGuildProgress(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.GetGuildProgressRequest
		mockSetup   func(storage *cloudsaveStorageMock, guildId string)
		expectedErr error
		expectedRes *pb.GuildProgress
	}{
		{
			name: "valid guild id",
			req: &pb.GetGuildProgressRequest{
				GuildId: "testId",
			},
			mockSetup: func(storage *cloudsaveStorageMock, guildId string) {
				storage.On("GetGuildProgress", "guildProgress_"+guildId).
					Return(&pb.GuildProgress{
						GuildId:    "testId",
						Objectives: map[string]int32{"testGoal": 1},
					}, nil)
			},
			expectedRes: &pb.GuildProgress{
				GuildId:    "testId",
				Objectives: map[string]int32{"testGoal": 1},
			},
		},
		{
			name: "invalid guild id",
			req: &pb.GetGuildProgressRequest{
				GuildId: "testId",
			},
			mockSetup: func(storage *cloudsaveStorageMock, guildId string) {
				storage.On("GetGuildProgress", "guildProgress_"+guildId).
					Return(&pb.GuildProgress{}, errors.New("error"))
			},
			expectedErr: status.Errorf(codes.Internal, "Error getting guild progress: error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokenRepo := mocks.NewMockTokenRepository(ctrl)
			refreshRepo := mocks.NewMockRefreshTokenRepository(ctrl)
			configRepo := mocks.NewMockConfigRepository(ctrl)
			storage := new(cloudsaveStorageMock)
			service := NewGuildServiceServer(tokenRepo, configRepo, refreshRepo, storage)
			tt.mockSetup(storage, tt.req.GuildId)

			// when
			res, err := service.GetGuildProgress(context.Background(), tt.req)

			// then
			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res.GuildProgress)
			}
			storage.AssertExpectations(t)
		})
	}
}
