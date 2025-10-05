package server

import (
	"context"
	"fmt"
	"strconv"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SecretHandler обрабатывает gRPC запросы для работы с секретами
type SecretHandler struct {
	v1.UnimplementedSecretServiceServer
	secretService *server.SecretService
}

// NewSecretHandler создает новый экземпляр SecretHandler
func NewSecretHandler(secretService *server.SecretService) *SecretHandler {
	return &SecretHandler{
		secretService: secretService,
	}
}

// CreateSecret создает новый секрет
func (h *SecretHandler) CreateSecret(ctx context.Context, req *v1.CreateSecretRequest) (*v1.CreateSecretResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := validateCreateSecretRequest(req); err != nil {
		return nil, err
	}

	secretType := convertProtoSecretType(req.GetType())
	if secretType == models.TypeUnspecified {
		return nil, status.Error(codes.InvalidArgument, "invalid secret type")
	}

	secret, err := h.secretService.CreateSecret(
		ctx,
		userID,
		secretType,
		req.GetName(),
		req.GetEncryptedData(),
		req.GetMetadata(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create secret: %v", err))
	}

	return &v1.CreateSecretResponse{
		SecretId:  strconv.Itoa(secret.ID),
		Version:   int32(secret.Version),
		CreatedAt: timestamppb.New(secret.CreatedAt),
	}, nil
}

// GetSecret возвращает секрет по ID
func (h *SecretHandler) GetSecret(ctx context.Context, req *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id is required")
	}

	secretID, err := strconv.Atoi(req.GetSecretId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid secret_id format")
	}

	secret, metadata, err := h.secretService.GetSecret(ctx, secretID, userID)
	if err != nil {
		return nil, mapSecretErrorToGRPC(err)
	}

	if secret == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	secretInfo := &v1.SecretInfo{
		Id:        req.GetSecretId(),
		Type:      convertModelsSecretType(secret.Type),
		Name:      secret.Name,
		Metadata:  metadata,
		Version:   int32(secret.Version),
		CreatedAt: timestamppb.New(secret.CreatedAt),
		UpdatedAt: timestamppb.New(secret.UpdatedAt),
	}

	if secret.DeletedAt != nil {
		secretInfo.DeletedAt = timestamppb.New(*secret.DeletedAt)
	}

	return &v1.GetSecretResponse{
		Info:          secretInfo,
		EncryptedData: secret.EncryptedData,
	}, nil
}

// UpdateSecret обновляет существующий секрет
func (h *SecretHandler) UpdateSecret(ctx context.Context, req *v1.UpdateSecretRequest) (*v1.UpdateSecretResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := validateUpdateSecretRequest(req); err != nil {
		return nil, err
	}

	secretID, err := strconv.Atoi(req.GetSecretId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid secret_id format")
	}

	updatedSecret, err := h.secretService.UpdateSecret(
		ctx,
		secretID,
		userID,
		req.GetName(),
		req.GetEncryptedData(),
		req.GetMetadata(),
		int(req.GetVersion()),
	)
	if err != nil {
		return nil, mapSecretErrorToGRPC(err)
	}

	return &v1.UpdateSecretResponse{
		NewVersion: int32(updatedSecret.Version),
		UpdatedAt:  timestamppb.New(updatedSecret.UpdatedAt),
	}, nil
}

// DeleteSecret помечает секрет как удаленный
func (h *SecretHandler) DeleteSecret(ctx context.Context, req *v1.DeleteSecretRequest) (*v1.DeleteSecretResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id is required")
	}

	secretID, err := strconv.Atoi(req.GetSecretId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid secret_id format")
	}

	err = h.secretService.DeleteSecret(ctx, secretID, userID)
	if err != nil {
		return nil, mapSecretErrorToGRPC(err)
	}

	return &v1.DeleteSecretResponse{
		Success:   true,
		DeletedAt: timestamppb.Now(),
	}, nil
}

// ListSecrets возвращает список секретов пользователя
func (h *SecretHandler) ListSecrets(ctx context.Context, req *v1.ListSecretsRequest) (*v1.ListSecretsResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	secrets, err := h.secretService.ListSecrets(ctx, userID)
	if err != nil {
		return nil, mapSecretErrorToGRPC(err)
	}

	secretInfos := make([]*v1.SecretInfo, 0, len(secrets))
	for _, secret := range secrets {
		if secret.DeletedAt != nil && !req.GetIncludeDeleted() {
			continue
		}

		var metadataWrapper models.MetadataWrapper
		if err := metadataWrapper.FromBytes(secret.Metadata); err != nil {
			continue
		}

		secretInfo := &v1.SecretInfo{
			Id:        strconv.Itoa(secret.ID),
			Type:      convertModelsSecretType(secret.Type),
			Name:      secret.Name,
			Metadata:  metadataWrapper.ToProto(),
			Version:   int32(secret.Version),
			CreatedAt: timestamppb.New(secret.CreatedAt),
			UpdatedAt: timestamppb.New(secret.UpdatedAt),
		}

		if secret.DeletedAt != nil {
			secretInfo.DeletedAt = timestamppb.New(*secret.DeletedAt)
		}

		secretInfos = append(secretInfos, secretInfo)
	}

	return &v1.ListSecretsResponse{
		Secrets: secretInfos,
	}, nil
}

// Sync синхронизирует секреты клиента с сервером
func (h *SecretHandler) Sync(ctx context.Context, req *v1.SyncRequest) (*v1.SyncResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	// Возвращаем все секреты пользователя - такая вот синхронизация
	secrets, err := h.secretService.ListSecrets(ctx, userID)
	if err != nil {
		return nil, mapSecretErrorToGRPC(err)
	}

	secretInfos := make([]*v1.SecretInfo, 0, len(secrets))
	for _, secret := range secrets {
		if secret.DeletedAt != nil {
			continue
		}

		var metadataWrapper models.MetadataWrapper
		if err := metadataWrapper.FromBytes(secret.Metadata); err != nil {
			continue
		}

		secretInfo := &v1.SecretInfo{
			Id:        strconv.Itoa(secret.ID),
			Type:      convertModelsSecretType(secret.Type),
			Name:      secret.Name,
			Metadata:  metadataWrapper.ToProto(),
			Version:   int32(secret.Version),
			CreatedAt: timestamppb.New(secret.CreatedAt),
			UpdatedAt: timestamppb.New(secret.UpdatedAt),
		}

		secretInfos = append(secretInfos, secretInfo)
	}

	return &v1.SyncResponse{
		Secrets: secretInfos,
		SyncState: &v1.SyncState{
			LastSync:         timestamppb.Now(),
			TotalSecrets:     int32(len(secretInfos)),
			UpdatedSecrets:   0,
			RequiresFullSync: false,
		},
	}, nil
}

func validateCreateSecretRequest(req *v1.CreateSecretRequest) error {
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetType() == v1.SecretType_SECRET_TYPE_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "secret type is required")
	}
	if len(req.GetEncryptedData()) == 0 {
		return status.Error(codes.InvalidArgument, "encrypted_data is required")
	}
	return nil
}

func validateUpdateSecretRequest(req *v1.UpdateSecretRequest) error {
	if req.GetSecretId() == "" {
		return status.Error(codes.InvalidArgument, "secret_id is required")
	}
	if req.GetName() == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetVersion() <= 0 {
		return status.Error(codes.InvalidArgument, "version must be positive")
	}
	return nil
}

func convertProtoSecretType(protoType v1.SecretType) models.SecretType {
	switch protoType {
	case v1.SecretType_SECRET_TYPE_LOGIN:
		return models.TypeLogin
	case v1.SecretType_SECRET_TYPE_CARD:
		return models.TypeCard
	case v1.SecretType_SECRET_TYPE_TEXT:
		return models.TypeText
	case v1.SecretType_SECRET_TYPE_BINARY:
		return models.TypeBinary
	default:
		return models.TypeUnspecified
	}
}

func convertModelsSecretType(modelType models.SecretType) v1.SecretType {
	switch modelType {
	case models.TypeLogin:
		return v1.SecretType_SECRET_TYPE_LOGIN
	case models.TypeCard:
		return v1.SecretType_SECRET_TYPE_CARD
	case models.TypeText:
		return v1.SecretType_SECRET_TYPE_TEXT
	case models.TypeBinary:
		return v1.SecretType_SECRET_TYPE_BINARY
	default:
		return v1.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

func mapSecretErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case err.Error() == "secret not found":
		return status.Error(codes.NotFound, "secret not found")
	case err.Error() == "version conflict":
		return status.Error(codes.FailedPrecondition, "version conflict")
	case err.Error() == "failed to update secret":
		return status.Error(codes.Internal, "failed to update secret")
	case err.Error() == "failed to delete secret":
		return status.Error(codes.Internal, "failed to delete secret")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
