package httputils

import (
	"context"
	"net/http"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	"github.com/lokeam/bravo-kilo/internal/shared/jwt"
	"github.com/lokeam/bravo-kilo/internal/shared/operations"
)

func AuthenticateRequest(
	ctx context.Context,
	r *http.Request,
	authExecutor *operations.OperationExecutor[int],
	) (int, error) {
	return authExecutor.Execute(ctx, func(ctx context.Context) (int, error) {
		userID, err := jwt.ExtractUserIDFromJWT(r, config.AppConfig.JWTPublicKey)
		if err != nil {
				return 0, core.ErrAuthentication
		}
		return userID, nil
	})
}