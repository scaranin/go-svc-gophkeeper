package main

import (
	"context"
	"fmt"
	"log"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	authClient := v1.NewAuthServiceClient(conn)
	secretClient := v1.NewSecretServiceClient(conn)

	ctx := context.Background()

	fmt.Println("01. Регистрация пользователя")
	registerResp, err := authClient.Register(ctx, &v1.RegisterRequest{
		Login:    "testuser",
		Password: "testderparol",
	})
	if err != nil {
		log.Printf("Register failed: %v", err)
	} else {
		fmt.Printf("Register successful: UserID=%s, Token=%s\n", registerResp.UserId, registerResp.AccessToken)
	}

	fmt.Println("02. Логин")
	loginResp, err := authClient.Login(ctx, &v1.LoginRequest{
		Login:    "testuser",
		Password: "testderparol",
	})
	if err != nil {
		log.Printf("Login failed: %v", err)
		return
	}
	fmt.Printf("Login successful: UserID=%s, Token=%s\n", loginResp.UserId, loginResp.AccessToken)

	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+loginResp.AccessToken,
	))

	fmt.Println("03. Добавляем секрет")
	createSecretResp, err := secretClient.CreateSecret(authCtx, &v1.CreateSecretRequest{
		Type: v1.SecretType_SECRET_TYPE_LOGIN,
		Name: "Test Login",
		Metadata: &v1.Metadata{
			Name:        "Яндекс почта",
			Description: "Пароль от почты",
			Website:     "https://mail.yandex.ru/",
			Tags:        "email",
		},
		EncryptedData: []byte("encrypted_login_data_here"),
	})
	if err != nil {
		log.Printf("CreateSecret failed: %v", err)
	} else {
		fmt.Printf("CreateSecret successful: SecretID=%s, Version=%d\n",
			createSecretResp.SecretId, createSecretResp.Version)
	}

	fmt.Println("04. Получить секреты")
	listResp, err := secretClient.ListSecrets(authCtx, &v1.ListSecretsRequest{
		IncludeDeleted: false,
	})
	if err != nil {
		log.Printf("ListSecrets failed: %v", err)
	} else {
		fmt.Printf("ListSecrets successful: Found %d secrets\n", len(listResp.Secrets))
		for i, secret := range listResp.Secrets {
			fmt.Printf("  %d. ID=%s, Type=%s, Name=%s, Version=%d\n",
				i+1, secret.Id, secret.Type, secret.Name, secret.Version)
		}
	}

	fmt.Println("05. Синхронизация")
	syncResp, err := secretClient.Sync(authCtx, &v1.SyncRequest{
		LastSync: nil,
	})
	if err != nil {
		log.Printf("Sync failed: %v", err)
	} else {
		fmt.Printf("Sync successful: %d secrets, LastSync=%v\n",
			syncResp.SyncState.TotalSecrets, syncResp.SyncState.LastSync.AsTime())
	}
}
