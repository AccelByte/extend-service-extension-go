// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
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

func createProtoFileDescriptor(s string) protoreflect.FileDescriptor {
	pb := new(descriptorpb.FileDescriptorProto)
	if err := prototext.Unmarshal([]byte(s), pb); err != nil {
		panic(err)
	}
	fd, err := protodesc.NewFile(pb, nil)
	if err != nil {
		panic(err)
	}

	return fd
}

func TestExtractPermission(t *testing.T) {
	tests := []struct {
		id                 string
		token              string
		protoPackageName   string
		protoServiceName   string
		protoMethodName    string
		protoMethodOptions string
		permission         *iam.Permission
		validateError      error
	}{
		{
			id:                 "no-permission-required",
			token:              "foo",
			protoPackageName:   "foo",
			protoServiceName:   "MyService",
			protoMethodName:    "MyMethod",
			protoMethodOptions: "",
			permission:         nil,
			validateError:      nil,
		},
		{
			id:               "permission-required",
			token:            "foo",
			protoPackageName: "service",
			protoServiceName: "Service",
			protoMethodName:  "CreateOrUpdateGuildProgress",
			protoMethodOptions: `options: {
				uninterpreted_option: [
					{
						name: [
							{
								name_part: "permission.action"
								is_extension: false
							}
						]
						positive_int_value: 1
					},
					{
						name: [
							{
								name_part: "permission.resource"
								is_extension: false
							}
						]
						string_value: "ADMIN:NAMESPACE:{namespace}:CLOUDSAVE:RECORD"
					}
				]
			}`,
			permission: &iam.Permission{
				Action:   1,
				Resource: "ADMIN:NAMESPACE:{namespace}:CLOUDSAVE:RECORD",
			},
			validateError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			namespace := getNamespace()
			val := &authValidatorMock{}
			val.On("Validate", test.token, test.permission, &namespace, mock.Anything).
				Return(test.validateError)
			Validator = val

			if _, err := protoregistry.GlobalFiles.FindFileByPath(fmt.Sprintf("%s.proto", test.protoPackageName)); err != nil {
				protoText := fmt.Sprintf(`syntax: "proto3"
name: "%[1]s.proto"
package: "%[1]s"
message_type: [
	{
		name: "Message"
	}
]
service: [
	{
		name: "%[2]s"
		method: [
			{
				name: "%[3]s"
				input_type: ".%[1]s.Message"
				output_type: ".%[1]s.Message"
				%[4]s
			}
		]
	}
]`,
					test.protoPackageName,
					test.protoServiceName,
					test.protoMethodName,
					test.protoMethodOptions,
				)
				protoFD := createProtoFileDescriptor(protoText)
				err := protoregistry.GlobalFiles.RegisterFile(protoFD)
				if err != nil {
					panic(err)
				}
			}

			md := map[string]string{
				"authorization": fmt.Sprintf("Bearer %s", test.token),
			}
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(md))
			req := struct{}{}
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return req, nil
			}
			info := &grpc.UnaryServerInfo{
				FullMethod: fmt.Sprintf("/%s.%s/%s", test.protoPackageName, test.protoServiceName, test.protoMethodName),
			}
			extractor := NewProtoPermissionExtractor()
			interceptor := NewUnaryAuthServerIntercept(extractor)
			_, err := interceptor(ctx, req, info, handler)
			assert.Equal(t, test.validateError, err)
		})
	}
}
