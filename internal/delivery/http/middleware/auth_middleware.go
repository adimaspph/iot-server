package middleware

import (
	"iot-server/internal/entity"
	"iot-server/internal/model"
	"iot-server/internal/usecase"
	"iot-server/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

const ctxAuthKey = "auth"

func NewAuth(userUC *usecase.UserUsecase, tokenUtil *util.TokenUtil) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				userUC.Log.Warn("missing Authorization header")
				return echo.ErrUnauthorized
			}

			// Accept "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				userUC.Log.Warn("invalid Authorization header format")
				return echo.ErrUnauthorized
			}

			token := strings.TrimSpace(parts[1])

			auth, err := tokenUtil.ParseToken(c.Request().Context(), token)
			if err != nil {
				userUC.Log.WithError(err).Warn("failed to parsing/validation token")
				return echo.ErrUnauthorized
			}

			// Store into Echo context for downstream handlers
			c.Set(ctxAuthKey, auth)
			return next(c)
		}
	}
}

func GetUser(c echo.Context) (*model.Auth, bool) {
	v := c.Get(ctxAuthKey)
	if v == nil {
		return nil, false
	}
	a, ok := v.(*model.Auth)
	return a, ok
}

func RequireRoles(allowed ...entity.Role) echo.MiddlewareFunc {
	allow := make(map[entity.Role]bool, len(allowed))
	for _, r := range allowed {
		allow[r] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, ok := GetUser(c)
			if !ok {
				return echo.ErrUnauthorized
			}
			if !allow[entity.Role(auth.Role)] {
				return echo.ErrForbidden
			}
			return next(c)
		}
	}
}
