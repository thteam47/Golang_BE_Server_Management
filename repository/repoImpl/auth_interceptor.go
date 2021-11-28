package repoimpl

import (
	"context"
	"fmt"
	"log"
	"strings"

	repo "github.com/thteam47/server_management/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtManager      repo.JwtRepository
	accessibleRoles map[string][]string
}

// NewAuthInterceptor returns a new auth interceptor
func NewAuthInterceptor(jwtManager repo.JwtRepository, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{jwtManager, accessibleRoles}
}

// Unary returns a server interceptor function to authenticate and authorize unary RPC
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		fmt.Println("--> unary interceptor: ", info.FullMethod)

		err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a server interceptor function to authenticate and authorize stream RPC
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		log.Println("--> stream interceptor: ", info.FullMethod)

		err := interceptor.authorize(stream.Context(), info.FullMethod)
		if err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
	accessibleRoles, ok := interceptor.accessibleRoles[method]
	if !ok {
		// everyone can access
		return nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) < 1 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	} else {

		accessToken := strings.TrimPrefix(values[0], "Bearer ")
		if accessToken == "undefined" {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}
		claims, err := interceptor.jwtManager.Verify(accessToken)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
		}

		for _, role := range accessibleRoles {
			if role == claims.Role {
				return nil
			}
			for _, action := range claims.Action {
				if role == action {
					return nil
				}
			}
		}
	}

	return status.Error(codes.PermissionDenied, "no permission to access this RPC")
}
