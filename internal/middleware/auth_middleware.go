package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/noedaka/go-url-shortener/internal/config"
)

const (
	cookieName = "session_token"
	secretKey  = "supersecretkey"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			userID := setNewCookie(w)
			ctx := context.WithValue(r.Context(), config.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			userID := setNewCookie(w)
			ctx := context.WithValue(r.Context(), config.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), config.UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setNewCookie(w http.ResponseWriter) string {
	userID := uuid.New().String()
	expiresAt := time.Now().Add(2 * time.Minute)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return userID
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return userID
}
