// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type authValidatorMock struct {
	mock.Mock
}

func (a *authValidatorMock) Initialize() {}
func (a *authValidatorMock) Validate(token string, permission *validator.Permission, namespace *string, userId *string) error {
	args := a.Called(token, permission, namespace, userId)

	return args.Error(0)
}

func TestUnaryAuthServerIntercept(t *testing.T) {
	md := map[string]string{
		"authorization": "Bearer <some-random-authorization-token>",
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(md))
	action := 2
	namespace := "test-accelbyte"
	resourceName := "test-CLOUDSAVE"
	perm := validator.Permission{
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

	// test
	res, err := UnaryAuthServerIntercept(ctx, req, nil, handler)
	assert.NoError(t, err)
	assert.Equal(t, req, res)
}
