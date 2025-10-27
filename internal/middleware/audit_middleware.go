package middleware

import (
    "context"
    "net/http"
    "time"

    "github.com/noedaka/go-url-shortener/internal/audit"
    "github.com/noedaka/go-url-shortener/internal/config"
    "github.com/noedaka/go-url-shortener/internal/model"
)

type auditKey struct{}

// AuditMiddleware добавляет менеджер аудита в контекст
func AuditMiddleware(auditManager *audit.AuditManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.WithValue(r.Context(), auditKey{}, auditManager)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GetAuditManagerFromContext возвращает менеджер аудита из контекста
func GetAuditManagerFromContext(ctx context.Context) *audit.AuditManager {
    if manager, ok := ctx.Value(auditKey{}).(*audit.AuditManager); ok {
        return manager
    }
    return nil
}

// LogAuditEvent логирует событие аудита
func LogAuditEvent(ctx context.Context, action, url string) {
    auditManager := GetAuditManagerFromContext(ctx)
    if auditManager == nil {
        return
    }

    userID, _ := ctx.Value(config.UserIDKey).(string)
    
    event := model.AuditEvent{
        TS:     time.Now().Unix(),
        Action: action,
        UserID: userID,
        URL:    url,
    }

    auditManager.NotifyObservers(event)
}