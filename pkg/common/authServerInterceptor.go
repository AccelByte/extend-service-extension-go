// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	pb "extend-custom-guild-service/pkg/pb"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	Validator validator.AuthTokenValidator
)

type ProtoPermissionExtractor interface {
	ExtractPermission(*grpc.UnaryServerInfo) (permission *validator.Permission, err error)
}

func NewProtoPermissionExtractor() *ProtoPermissionExtractorImpl {
	return &ProtoPermissionExtractorImpl{}
}

type ProtoPermissionExtractorImpl struct{}

func (p *ProtoPermissionExtractorImpl) ExtractPermission(info *grpc.UnaryServerInfo) (*validator.Permission, error) {
	serviceName, methodName, err := parseFullMethod(info.FullMethod)
	if err != nil {
		return nil, err
	}

	// Read the required permission stated in the proto file
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(serviceName))
	if err != nil {
		return nil, err
	}

	serviceDesc := desc.(protoreflect.ServiceDescriptor)
	method := serviceDesc.Methods().ByName(protoreflect.Name(methodName))
	resource := proto.GetExtension(method.Options(), pb.E_Resource).(string)
	action := proto.GetExtension(method.Options(), pb.E_Action).(pb.Action)
	permission := wrapPermission(resource, int(action.Number()))

	return &permission, nil
}

func NewUnaryAuthServerIntercept(
	permissionExtractor ProtoPermissionExtractor,
) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) { // nolint

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if Validator == nil {
			return nil, errors.New("server token validator not set")
		}

		meta, found := metadata.FromIncomingContext(ctx)
		if !found {
			return nil, errors.New("metadata missing")
		}

		authorization := meta["authorization"][0]
		token := strings.TrimPrefix(authorization, "Bearer ")

		// Extract permission stated in the proto file
		permission, err := permissionExtractor.ExtractPermission(info)
		if err != nil {
			return nil, err
		}
		namespace := getNamespace()

		err = Validator.Validate(token, permission, &namespace, nil)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func parseFullMethod(fullMethod string) (string, string, error) {
	// Define the regular expression according to example shown here https://github.com/grpc/grpc-java/issues/4726
	re := regexp.MustCompile(`^/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(fullMethod)

	// Validate the match
	if matches == nil {
		return "", "", fmt.Errorf("invalid FullMethod format")
	}

	// Extract service and method names
	serviceName, methodName := matches[1], matches[2]

	if len(serviceName) == 0 {
		return "", "", fmt.Errorf("invalid FullMethod format: service name is empty")
	}

	if len(methodName) == 0 {
		return "", "", fmt.Errorf("invalid FullMethod format: method name is empty")
	}

	return serviceName, methodName, nil
}

func StreamAuthServerIntercept(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if Validator == nil {
		return errors.New("server token validator not set")
	}

	meta, found := metadata.FromIncomingContext(ss.Context())
	if !found {
		return errors.New("metadata missing")
	}

	authorization := meta["authorization"][0]
	token := strings.TrimPrefix(authorization, "Bearer ")

	namespace := getNamespace()
	permission := getRequiredPermission()
	var userId *string

	err := Validator.Validate(token, &permission, &namespace, userId)
	if err != nil {
		return err
	}

	return handler(srv, ss)
}

func getAction() int {
	return GetEnvInt("AB_ACTION", 2)
}

func getNamespace() string {
	return GetEnv("AB_NAMESPACE", "accelbyte")
}

func getResourceName() string {
	return GetEnv("AB_RESOURCE_NAME", "CLOUDSAVE")
}

func wrapPermission(resource string, action int) validator.Permission {
	return validator.Permission{
		Action:   action,
		Resource: resource,
	}
}

func getRequiredPermission() validator.Permission {
	return validator.Permission{
		Action:   getAction(),
		Resource: fmt.Sprintf("NAMESPACE:%s:%s", getNamespace(), getResourceName()),
	}
}
