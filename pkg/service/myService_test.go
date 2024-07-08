// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.
package service

import (
	"context"
	"errors"
	pb "extend-custom-guild-service/pkg/pb"
	"extend-custom-guild-service/pkg/service/mocks"
	"fmt"
	"testing"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"time"
)

//go:generate mockgen -destination ./mocks/server_mock.go -package mocks extend-custom-guild-service/pkg/pb myServiceServer
//go:generate mockgen -destination ./mocks/repo_mock.go -package mocks github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository TokenRepository,ConfigRepository,RefreshTokenRepository

type cloudsaveStorageMock struct {
	mock.Mock
}

func (c *cloudsaveStorageMock) GetGuildProgress(namespace string, key string) (*pb.GuildProgress, error) {
	args := c.Called(namespace, key)

	return args.Get(0).(*pb.GuildProgress), args.Error(1)
}

func (c *cloudsaveStorageMock) SaveGuildProgress(namespace string, key string, value *pb.GuildProgress) (*pb.GuildProgress, error) {
	args := c.Called(namespace, key, value)

	return args.Get(0).(*pb.GuildProgress), args.Error(1)
}

func TestMyServiceServerImpl_CreateOrUpdateGuildProgress(t *testing.T) {
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
				Namespace: "testNamespace",
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
				Namespace: "testNamespace",
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
			service := NewMyServiceServer(tokenRepo, configRepo, refreshRepo, storage)

			namespace := "testNamespace"
			guildProgressKey := fmt.Sprintf("guildProgress_%s", tt.req.GuildProgress.GuildId)
			storage.On("SaveGuildProgress", namespace, guildProgressKey, tt.req.GuildProgress).Return(tt.req.GetGuildProgress(), tt.expectedErr)

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

func TestMyServiceServerImpl_GetGuildProgress(t *testing.T) {
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
				Namespace: "testNamespace",
				GuildId:   "testId",
			},
			mockSetup: func(storage *cloudsaveStorageMock, guildId string) {
				storage.On("GetGuildProgress", "testNamespace", "guildProgress_"+guildId).
					Return(&pb.GuildProgress{
						GuildId:    "testId",
						Namespace:  "testNamespace",
						Objectives: map[string]int32{"testGoal": 1},
					}, nil)
			},
			expectedRes: &pb.GuildProgress{
				GuildId:    "testId",
				Namespace:  "testNamespace",
				Objectives: map[string]int32{"testGoal": 1},
			},
		},
		{
			name: "invalid guild id",
			req: &pb.GetGuildProgressRequest{
				Namespace: "testNamespace",
				GuildId:   "testId",
			},
			mockSetup: func(storage *cloudsaveStorageMock, guildId string) {
				storage.On("GetGuildProgress", "testNamespace", "guildProgress_"+guildId).
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
			service := NewMyServiceServer(tokenRepo, configRepo, refreshRepo, storage)
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
func TestMyServiceServerImpl_With_Refresh(t *testing.T) {
	tests := []struct {
		name            string
		req             *pb.CreateOrUpdateGuildProgressRequest
		wantErr         bool
		expectedErr     error
		expectedGuildId string
	}{
		{
			name: "successful create guild with refresh token",
			req: &pb.CreateOrUpdateGuildProgressRequest{
				Namespace: "testNamespace",
				GuildProgress: &pb.GuildProgress{
					GuildId:    "testId",
					Objectives: map[string]int32{"testGoal": 1},
				},
			},
			wantErr:         false,
			expectedGuildId: "testId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tokenRepo := mocks.NewMockTokenRepository(ctrl)
			refreshRepo := mocks.NewMockRefreshTokenRepository(ctrl)
			configRepo := mocks.NewMockConfigRepository(ctrl)
			storage := new(cloudsaveStorageMock)

			oauthService := iam.OAuth20Service{
				TokenRepository:        tokenRepo,
				RefreshTokenRepository: refreshRepo,
				ConfigRepository:       configRepo,
			}
			mocks.SetupTokenRepositoryExpectations(tokenRepo)
			mocks.SetupRefreshTokenRepositoryExpectations(refreshRepo)

			getToken, err := oauthService.TokenRepository.GetToken()
			require.NoError(t, err)

			Repository := oauthService.GetAuthSession().Refresh

			t.Logf("token is expected to expire in... : %v", repository.GetSecondsTillExpiry(tokenRepo, Repository.GetRefreshRate()))

			// Force the Token to be expired
			expiresIn := int32(5)
			mocks.MonkeyPatchTokenExpiry(getToken, expiresIn)

			getExpiresIn, err := repository.GetExpiresIn(oauthService.TokenRepository)
			require.NoError(t, err)
			t.Logf("monkey patched expiring in %vs", *getExpiresIn)

			secondsTillExpiry := repository.GetSecondsTillExpiry(oauthService.TokenRepository, Repository.GetRefreshRate())
			t.Logf("token is forced to expire in... : %v", secondsTillExpiry)

			errStore := oauthService.TokenRepository.Store(*getToken) // store the new monkey-patched Token
			require.NoError(t, errStore)

			sleepUntilTokenExpires(getExpiresIn)

			hasTokenExpired := repository.HasTokenExpired(oauthService.TokenRepository, Repository.GetRefreshRate())
			assert.True(t, hasTokenExpired, "token should be expired.") // Token expired
			t.Log("token has expired and refreshed.")

			// token is renewed and call the service
			service := NewMyServiceServer(tokenRepo, configRepo, refreshRepo, storage)

			namespace := "testNamespace"
			guildProgressKey := fmt.Sprintf("guildProgress_%s", tt.req.GuildProgress.GuildId)
			storage.On("SaveGuildProgress", namespace, guildProgressKey, tt.req.GuildProgress).Return(tt.req.GetGuildProgress(), tt.expectedErr)

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

func sleepUntilTokenExpires(getExpiresIn *int32) {
	tdu := time.Duration(*getExpiresIn) * time.Second
	logrus.Printf("sleep for %v second...", tdu)
	time.Sleep(tdu)
}
