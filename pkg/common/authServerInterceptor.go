// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package common

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	pb "extend-custom-guild-service/pkg/pb"
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/AccelByte/accelbyte-go-sdk/iam-sdk/pkg/iamclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/iam"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/utils/auth/validator"
	"github.com/pkg/errors"
)

var (
	Validator validator.AuthTokenValidator
)

type ProtoPermissionExtractor interface {
	ExtractPermission(infoUnary *grpc.UnaryServerInfo, infoStream *grpc.StreamServerInfo) (permission *validator.Permission, err error)
}

func NewProtoPermissionExtractor() *ProtoPermissionExtractorImpl {
	return &ProtoPermissionExtractorImpl{}
}

type ProtoPermissionExtractorImpl struct{}

func (p *ProtoPermissionExtractorImpl) ExtractPermission(infoUnary *grpc.UnaryServerInfo, infoStream *grpc.StreamServerInfo) (*validator.Permission, error) {
	if infoUnary != nil && infoStream != nil {
		return nil, errors.New("both infoUnary and infoStream cannot be filled at the same time")
	}

	var serviceName string
	var methodName string
	var err error

	if infoUnary != nil {
		serviceName, methodName, err = parseFullMethod(infoUnary.FullMethod)
	} else if infoStream != nil {
		serviceName, methodName, err = parseFullMethod(infoStream.FullMethod)
	} else {
		return nil, errors.New("both infoUnary and infoStream are nil")
	}
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
		permission, err := permissionExtractor.ExtractPermission(info, nil)
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

func NewStreamAuthServerIntercept(
	permissionExtractor ProtoPermissionExtractor,
) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if strings.Contains(info.FullMethod, "grpc.reflection") {
			return handler(srv, ss)
		}

		if Validator == nil {
			return errors.New("server token validator not set")
		}

		meta, found := metadata.FromIncomingContext(ss.Context())
		if !found {
			return errors.New("metadata missing")
		}

		authorization := meta["authorization"][0]
		token := strings.TrimPrefix(authorization, "Bearer ")

		// Extract permission stated in the proto file
		permission, err := permissionExtractor.ExtractPermission(nil, info)
		if err != nil {
			return err
		}
		namespace := getNamespace()

		err = Validator.Validate(token, permission, &namespace, nil)
		if err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

func getNamespace() string {
	return GetEnv("AB_NAMESPACE", "accelbyte")
}

func wrapPermission(resource string, action int) validator.Permission {
	return validator.Permission{
		Action:   action,
		Resource: resource,
	}
}

func NewTokenValidator(authService iam.OAuth20Service, refreshInterval time.Duration, validateLocally bool) validator.AuthTokenValidator {
	return &validator.TokenValidator{
		AuthService:     authService,
		RefreshInterval: refreshInterval,

		Filter:                nil,
		JwkSet:                nil,
		JwtClaims:             validator.JWTClaims{},
		JwtEncoding:           *base64.URLEncoding.WithPadding(base64.NoPadding),
		PublicKeys:            make(map[string]*rsa.PublicKey),
		LocalValidationActive: validateLocally,
		RevokedUsers:          make(map[string]time.Time),
		Roles:                 make(map[string]*iamclientmodels.ModelRoleResponseV3),
	}
}
