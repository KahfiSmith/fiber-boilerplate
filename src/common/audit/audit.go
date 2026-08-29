package audit

import (
	"log/slog"
	"time"
)

type EventType string

const (
	EventLoginSuccess           EventType = "LOGIN_SUCCESS"
	EventLoginFailed            EventType = "LOGIN_FAILED"
	EventLogout                 EventType = "LOGOUT"
	EventPasswordChanged        EventType = "PASSWORD_CHANGED"
	EventPasswordResetRequested EventType = "PASSWORD_RESET_REQUESTED"
	EventPasswordResetCompleted EventType = "PASSWORD_RESET_COMPLETED"
	EventEmailVerified          EventType = "EMAIL_VERIFIED"
	EventSessionRevoked         EventType = "SESSION_REVOKED"
	EventRefreshTokenReused     EventType = "REFRESH_TOKEN_REUSED"
	EventAccountDeleted         EventType = "ACCOUNT_DELETED"
)

func LogEvent(eventType EventType, userID uint, sessionID string, ip string, userAgent string, details map[string]interface{}) {
	attrs := []slog.Attr{
		slog.String("audit_event", string(eventType)),
		slog.Time("timestamp", time.Now()),
	}

	if userID != 0 {
		attrs = append(attrs, slog.Uint64("user_id", uint64(userID)))
	}
	if sessionID != "" {
		attrs = append(attrs, slog.String("session_id", sessionID))
	}
	if ip != "" {
		attrs = append(attrs, slog.String("ip_address", ip))
	}
	if userAgent != "" {
		attrs = append(attrs, slog.String("user_agent", userAgent))
	}

	for k, v := range details {
		attrs = append(attrs, slog.Any(k, v))
	}

	slog.Default().LogAttrs(nil, slog.LevelInfo, "SECURITY_AUDIT", attrs...)
}
