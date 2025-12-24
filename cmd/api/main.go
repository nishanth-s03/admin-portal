package main

import (
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"github.com/joho/godotenv"

	authpb "admin-portal/proto/auth"

	"admin-portal/internal/auth-module/handler"
	"admin-portal/internal/auth-module/middleware"
	"admin-portal/internal/auth-module/repository"
	"admin-portal/internal/auth-module/service"

	"admin-portal/internal/shared/database"
	"admin-portal/internal/shared/security"
)

func main() {
	log.Println("🚀 API server starting...")
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Failed to load .env:", err)
	}
	// ---------------------------
	// Load DB & GORM
	// ---------------------------
	cfg := database.LoadConfig()

	db, err := database.OpenGorm(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// ---------------------------
	// Initialize repositories
	// ---------------------------
	userRepo := repository.NewUserRepository(db)
	passwordRepo := repository.NewPasswordRepository(db)
	loginLogRepo := repository.NewLoginLogRepository(db)
	userSessionRepo := repository.NewUserSessionRepository(db)

	// ---------------------------
	// JWT configuration
	// ---------------------------
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	jwtCfg := security.JWTConfig{
		Secret:           JWTSecret,
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		Issuer:           "admin-portal",
	}

	// ---------------------------
	// Initialize services
	// ---------------------------
	tokenService := service.NewTokenService(jwtCfg, userSessionRepo)

	authService := service.NewAuthService(
		db,
		userRepo,
		passwordRepo,
		loginLogRepo,
		tokenService,
	)

	// ---------------------------
	// Initialize handlers
	// ---------------------------
	authHandler := handler.NewAuthHandler(authService)

	// ---------------------------
	// gRPC server with interceptors
	// ---------------------------
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.JWTUnaryInterceptor(jwtCfg),
		),
	)

	// ---------------------------
	// Register gRPC services
	// ---------------------------
	authpb.RegisterAuthServiceServer(grpcServer, authHandler)

	// ---------------------------
	// Start server
	// ---------------------------
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("🚀 gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
