// internal/presentation/http/middlewares/auth_middleware.go
package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/farmanexo/user-service/internal/infrastructure/security"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDCtxKey      contextKey = "user_id"
	UserRoleCtxKey    contextKey = "user_role"
	AccessTokenCtxKey contextKey = "access_token"
)

// AuthMiddleware maneja la autenticación JWT en rutas protegidas
type AuthMiddleware struct {
	jwtService security.JWTService
	logger     *zap.Logger
}

func NewAuthMiddleware(jwtService security.JWTService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

// RequireAuth middleware que valida el access token JWT
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Request sin header Authorization",
				zap.String("path", r.URL.Path),
			)
			m.respondUnauthorized(w, "Header Authorization es requerido")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.respondUnauthorized(w, "Formato de token inválido. Use: Bearer {token}")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			m.respondUnauthorized(w, "Token vacío")
			return
		}

		userID, role, _, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			m.logger.Warn("Access token inválido",
				zap.String("path", r.URL.Path),
				zap.Error(err),
			)
			m.respondUnauthorized(w, "Token inválido o expirado")
			return
		}

		m.logger.Debug("Token validado exitosamente",
			zap.String("user_id", userID),
			zap.String("role", role),
			zap.String("path", r.URL.Path),
		)

		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		ctx = context.WithValue(ctx, UserRoleCtxKey, role)
		ctx = context.WithValue(ctx, AccessTokenCtxKey, tokenString)
		ctx = mediator.WithValue(ctx, mediator.UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	response := common.NewApiResponse[struct{}]()
	response.SetHttpStatus(constants.StatusUnauthorized.Int())
	response.AddError(constants.CodeUnauthorized, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(response)
}

// GetUserIDFromContext obtiene el user ID del contexto
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDCtxKey).(string)
	return userID, ok
}

// GetUserRoleFromContext obtiene el role del contexto
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleCtxKey).(string)
	return role, ok
}
