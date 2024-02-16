// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"google.golang.org/grpc"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type authValidatorMock struct {
	mock.Mock
}

func (a *authValidatorMock) Initialize() {}
func (a *authValidatorMock) Validate(token string, permission *iam.Permission, namespace *string, userId *string) error {
	args := a.Called(token, permission, namespace, userId)

	return args.Error(0)
}

type permissionExtractorMock struct {
	mock.Mock
}

func (p *permissionExtractorMock) ExtractPermission(infoUnary *grpc.UnaryServerInfo, infoStream *grpc.StreamServerInfo) (permission *iam.Permission, err error) {
	args := p.Called(infoUnary)

	return args.Get(0).(*iam.Permission), args.Error(1)
}

func TestUnaryAuthServerIntercept(t *testing.T) {
	md := map[string]string{
		"authorization": "Bearer <some-random-authorization-token>",
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(md))
	action := 2
	namespace := "test-accelbyte"
	resourceName := "test-CLOUDSAVE"
	perm := iam.Permission{
		Action:   action,
		Resource: fmt.Sprintf("NAMESPACE:%s:%s", namespace, resourceName),
	}
	var userId *string
	t.Setenv("AB_ACTION", strconv.Itoa(action))
	t.Setenv("AB_NAMESPACE", namespace)
	t.Setenv("AB_RESOURCE_NAME", resourceName)

	val := &authValidatorMock{}
	val.On("Validate", "<some-random-authorization-token>", &perm, &namespace, userId).Return(nil)
	Validator = val

	req := struct{}{}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/abc.def.MyService/MyMethod",
	}
	extractor := &permissionExtractorMock{}
	extractor.On("ExtractPermission", info).Return(&perm, nil)

	// test
	interceptor := NewUnaryAuthServerIntercept(extractor)
	res, err := interceptor(ctx, req, info, handler)
	assert.NoError(t, err)
	assert.Equal(t, req, res)
}
