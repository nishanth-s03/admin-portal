package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "admin-portal/proto/auth"
)

func main() {
	conn, err := grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := authpb.NewAuthServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// -------- Register --------
	log.Println("➡ Registering user...")
	regResp, err := client.Register(ctx, &authpb.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Role:     "admin",
	})
	if err != nil {
		log.Fatal("Register failed:", err)
	}
	log.Println("✅ Registered user:", regResp.UserId)

	// --------Activate User --------
	log.Println("➡ Activating user...")
	_, err = client.Activate(ctx, &authpb.ActivateRequest{
		UserId: regResp.UserId,
	})
	if err != nil {
		log.Fatal("Activate user failed:", err)
	}
	log.Println("✅ Activated user:", regResp.UserId)

	// -------- Login --------
	log.Println("➡ Logging in...")
	loginResp, err := client.Login(ctx, &authpb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		log.Fatal("Login failed:", err)
	}
	log.Println("✅ Logged in user:", loginResp.UserId)

	// -------- Logout --------
	log.Println("➡ Logging out...")
	_, err = client.Logout(ctx, nil)
	if err != nil {
		log.Fatal("Logout failed:", err)
	}
	log.Println("✅ Logged out successfully")
}
