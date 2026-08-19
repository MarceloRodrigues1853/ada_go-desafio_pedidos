package logger

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type contextKey string

const sagaIDKey contextKey = "saga_id"

// WithSagaID adiciona o saga_id ao contexto
func WithSagaID(ctx context.Context, sagaID uuid.UUID) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, sagaIDKey, sagaID)
}

// SagaIDFromContext extrai o saga_id do contexto, se presente
func SagaIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	val, ok := ctx.Value(sagaIDKey).(uuid.UUID)
	if !ok {
		if s, ok := ctx.Value(sagaIDKey).(string); ok {
			if parsed, err := uuid.Parse(s); err == nil {
				return parsed, true
			}
		}
		return uuid.Nil, false
	}
	return val, true
}

// ContextHandler intercepta logs contextuais para injetar saga_id automaticamente
type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{Handler: h}
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if sagaID, ok := SagaIDFromContext(ctx); ok {
			hasSaga := false
			r.Attrs(func(a slog.Attr) bool {
				if a.Key == "saga_id" {
					hasSaga = true
					return false
				}
				return true
			})
			if !hasSaga {
				r.AddAttrs(slog.String("saga_id", sagaID.String()))
			}
		}
	}
	return h.Handler.Handle(ctx, r)
}
