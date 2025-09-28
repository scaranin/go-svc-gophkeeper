package server

import (
	"context"
	"strings"

	"go-svc-gophkeeper/internal/app/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTAuthMiddleware создает gRPC interceptor для JWT аутентификации
func JWTAuthMiddleware(authService *server.AuthService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		userID, err := extractUserIDFromToken(ctx, authService)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, userIDKey{}, userID)
		return handler(ctx, req)
	}
}

// userIDKey тип для ключа в context
type userIDKey struct{}

// isPublicMethod проверяет, является ли метод публичным и не требует аутентификации
func isPublicMethod(fullMethod string) bool {
	publicMethods := map[string]bool{
		"/gophkeeper.v1.AuthService/Login":    true,
		"/gophkeeper.v1.AuthService/Register": true,
	}

	return publicMethods[fullMethod]
}

// extractUserIDFromToken извлекает userID из JWT токена в заголовках
func extractUserIDFromToken(ctx context.Context, authService *server.AuthService) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := strings.TrimSpace(authHeaders[0])
	if !strings.HasPrefix(token, "Bearer ") {
		return 0, status.Error(codes.Unauthenticated, "invalid authorization format, expected 'Bearer <token>'")
	}

	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return 0, status.Error(codes.Unauthenticated, "empty token")
	}

	userID, err := authService.ValidateToken(token)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	return userID, nil
}

// GetUserIDFromContext вспомогательная функция для извлечения userID из context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey{}).(int)
	return userID, ok
}
