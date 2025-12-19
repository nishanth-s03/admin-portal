package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/golang-jwt/jwt/v5"

	"admin-portal/internal/shared/security"
)

func JWTUnaryInterceptor(
	cfg security.JWTConfig,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Allow unauthenticated endpoints
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		tokenStr, err := extractTokenFromMetadata(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing auth token")
		}

		claims, err := validateJWT(cfg, tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Inject user into context
		ctx = WithAuthContext(
			ctx,
			claims.UserID,
			claims.Username,
			claims.Role,
		)

		return handler(ctx, req)
	}
}

func validateJWT(
	cfg security.JWTConfig,
	tokenStr string,
) (*security.Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&security.Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*security.Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	// 1️⃣ From Authorization header: "Bearer <token>"
	if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
		parts := strings.SplitN(authHeaders[0], " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1], nil
		}
	}

	// 2️⃣ From cookie header
	if cookies := md.Get("cookie"); len(cookies) > 0 {
		for _, c := range cookies {
			for _, part := range strings.Split(c, ";") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "access_token=") {
					return strings.TrimPrefix(part, "access_token="), nil
				}
			}
		}
	}

	return "", status.Error(codes.Unauthenticated, "token not found")
}

func isPublicMethod(method string) bool {
	switch method {
	case "/auth.AuthService/Login",
		"/auth.AuthService/Register",
		"/auth.AuthService/Activate":
		return true
	default:
		return false
	}
}
