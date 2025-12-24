package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	authpb "admin-portal/proto/auth"
	"admin-portal/internal/auth-module/service"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(
	ctx context.Context,
	req *authpb.RegisterRequest,
) (*authpb.RegisterResponse, error) {

	user, err := h.authService.Register(
		ctx,
		req.GetUsername(),
		req.GetPassword(),
		req.GetRole(),
	)
	if err != nil {
		return nil, err
	}

	return &authpb.RegisterResponse{
		UserId: user.ID.String(),
	}, nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *authpb.LoginRequest,
) (*authpb.LoginResponse, error) {

	user, accessToken, refreshToken, err :=
		h.authService.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	// Access token cookie
	grpc.SendHeader(ctx, metadata.Pairs(
		"set-cookie",
		buildCookie(
			"access_token",
			accessToken,
			"/",
			true,
		),
	))

	// Refresh token cookie
	grpc.SendHeader(ctx, metadata.Pairs(
		"set-cookie",
		buildCookie(
			"refresh_token",
			refreshToken,
			"/auth/refresh",
			true,
		),
	))

	return &authpb.LoginResponse{
		UserId: user.ID.String(),
	}, nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	_ *emptypb.Empty,
) (*emptypb.Empty, error) {

	// Extract refresh token from cookie
	md, _ := metadata.FromIncomingContext(ctx)
	for _, c := range md.Get("cookie") {
		if token := extractCookie(c, "refresh_token"); token != "" {
			_ = h.authService.Logout(ctx, token)
		}
	}

	// Clear cookies
	grpc.SendHeader(ctx, metadata.Pairs(
		"set-cookie", clearCookie("access_token", "/"),
	))
	grpc.SendHeader(ctx, metadata.Pairs(
		"set-cookie", clearCookie("refresh_token", "/auth/refresh"),
	))

	return &emptypb.Empty{}, nil
}

func (h *AuthHandler) Activate(
	ctx context.Context,
	req *authpb.ActivateRequest,
) (*emptypb.Empty, error) {

	if err := h.authService.ActivateUser(ctx, req.GetUserId()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

//-------------------- Helper functions for cookie management --------------------//

func buildCookie(
	name string,
	value string,
	path string,
	httpOnly bool,
) string {
	parts := []string{
		name + "=" + value,
		"Path=" + path,
		"SameSite=Strict",
	}

	if httpOnly {
		parts = append(parts, "HttpOnly")
	}

	parts = append(parts, "Secure")
	return strings.Join(parts, "; ")
}

func clearCookie(name, path string) string {
	return name + "=; Path=" + path + "; Max-Age=0; HttpOnly; Secure; SameSite=Strict"
}

func extractCookie(header string, name string) string {
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+"=") {
			return strings.TrimPrefix(part, name+"=")
		}
	}
	return ""
}
