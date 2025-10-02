package errors

// AppError представляет упрощенную ошибку приложения
type AppError struct {
	Code    string
	Message string
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

// Создаем типизированные ошибки как константы
var (
	// Общие ошибки
	ErrInternal      = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error"}
	ErrInvalidInput  = &AppError{Code: "INVALID_INPUT", Message: "invalid input"}
	ErrNotFound      = &AppError{Code: "NOT_FOUND", Message: "not found"}
	ErrAlreadyExists = &AppError{Code: "ALREADY_EXISTS", Message: "already exists"}
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Message: "unauthorized"}
	ErrForbidden     = &AppError{Code: "FORBIDDEN", Message: "forbidden"}
	ErrConflict      = &AppError{Code: "CONFLICT", Message: "conflict"}

	// Ошибки аутентификации
	ErrCredentialsRequired = &AppError{Code: "REQUIRED_CREDENTIALS", Message: "login and password required"}
	ErrInvalidCredentials  = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials"}
	ErrTokenExpired        = &AppError{Code: "TOKEN_EXPIRED", Message: "token expired"}
	ErrTokenInvalid        = &AppError{Code: "TOKEN_INVALID", Message: "token invalid"}

	// Ошибки валидации
	ErrValidation = &AppError{Code: "VALIDATION_ERROR", Message: "validation failed"}

	// Ошибки шифрования
	ErrEncryptionFailed = &AppError{Code: "ENCRYPTION_FAILED", Message: "encryption failed"}
	ErrDecryptionFailed = &AppError{Code: "DECRYPTION_FAILED", Message: "decryption failed"}

	// Ошибки базы данных
	ErrDBConnection = &AppError{Code: "DB_CONNECTION_ERROR", Message: "database connection failed"}
	ErrDBQuery      = &AppError{Code: "DB_QUERY_ERROR", Message: "database query failed"}
)

// Is проверяет, является ли ошибка определенного типа
func Is(err error, target *AppError) bool {
	if err == nil || target == nil {
		return false
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == target.Code
	}

	return false
}

// New создает новую ошибку с кодом и сообщением
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
