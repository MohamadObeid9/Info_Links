package api

import (
	"context"
	"log/slog"

	"infolinks-backend/internal/models"
)

type Handler struct {
	logger        *slog.Logger
	jwtSecret     []byte
	reportService reportService
}

type Dependencies struct {
	Logger        *slog.Logger
	JWTSecret     []byte
	ReportService reportService
}

type reportService interface {
	Create(ctx context.Context, report models.Report) error
}

func NewHandler(deps Dependencies) *Handler {
	handlerHandler := Handler{
		logger:    deps.Logger,
		jwtSecret: deps.JWTSecret,
	}
	return &handlerHandler
}
