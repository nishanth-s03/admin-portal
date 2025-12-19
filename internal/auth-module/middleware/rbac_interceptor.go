package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RBACUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		role, ok := RoleFromContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "role not found")
		}

		_ = role // enforce rules later

		return handler(ctx, req)
	}
}
