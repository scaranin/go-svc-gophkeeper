package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Адрес gRPC сервера")
	login := flag.String("login", "testuser", "Логин пользователя")
	password := flag.String("password", "testderparol", "Пароль пользователя")
	flag.Parse()

	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	authClient := v1.NewAuthServiceClient(conn)
	secretClient := v1.NewSecretServiceClient(conn)

	ctx := context.Background()

	fmt.Println("01. Регистрация пользователя")
	registerResp, err := authClient.Register(ctx, &v1.RegisterRequest{
		Login:    *login,
		Password: *password,
	})
	if err != nil {
		log.Printf("Ошибка регистрации: %v", err)
	} else {
		fmt.Printf("Регистрация успешна: UserID=%s, Token=%s\n", registerResp.UserId, registerResp.AccessToken)
	}

	fmt.Println("02. Вход в систему")
	loginResp, err := authClient.Login(ctx, &v1.LoginRequest{
		Login:    *login,
		Password: *password,
	})
	if err != nil {
		log.Printf("Ошибка входа: %v", err)
		return
	}
	fmt.Printf("Вход выполнен успешно: UserID=%s, Token=%s\n", loginResp.UserId, loginResp.AccessToken)

	authCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+loginResp.AccessToken,
	))

	fmt.Println("03. Добавление секрета")
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
		log.Printf("Ошибка создания секрета: %v", err)
	} else {
		fmt.Printf("Секрет создан успешно: SecretID=%s, Версия=%d\n",
			createSecretResp.SecretId, createSecretResp.Version)
	}

	fmt.Println("04. Получение списка секретов")
	listResp, err := secretClient.ListSecrets(authCtx, &v1.ListSecretsRequest{
		IncludeDeleted: false,
	})
	if err != nil {
		log.Printf("Ошибка получения списка секретов: %v", err)
	} else {
		fmt.Printf("Список секретов получен успешно: найдено %d секретов\n", len(listResp.Secrets))
		for i, secret := range listResp.Secrets {
			fmt.Printf("  %d. ID=%s, Тип=%s, Имя=%s, Версия=%d\n",
				i+1, secret.Id, secret.Type, secret.Name, secret.Version)
		}
	}

	fmt.Println("05. Синхронизация")
	syncResp, err := secretClient.Sync(authCtx, &v1.SyncRequest{
		LastSync: nil,
	})
	if err != nil {
		log.Printf("Ошибка синхронизации: %v", err)
	} else {
		fmt.Printf("Синхронизация выполнена успешно: %d секретов, Последняя синхронизация=%v\n",
			syncResp.SyncState.TotalSecrets, syncResp.SyncState.LastSync.AsTime())
	}
}
