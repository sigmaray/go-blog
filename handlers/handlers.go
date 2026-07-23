package handlers

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Handler holds shared dependencies for HTTP handlers.
// DB is the GORM database handle used by route handlers.
// Validate is the shared request validator instance.
type Handler struct {
	DB       *gorm.DB
	Validate *validator.Validate
}

// NewHandler constructs a Handler with the given database connection.
// db is the GORM handle injected into every handler method.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:       db,
		Validate: validator.New(),
	}
}
