package middleware

import (
	"github.com/K-H-Tech/apm/internal/contract"
)

// Middlewares will implement all required ad necessary middlewares for the http server
type Middlewares struct {
	Logger contract.Logger
}

// NewMiddlewares is the Middlewares factory method
func NewMiddlewares(logger contract.Logger) *Middlewares {
	return &Middlewares{
		Logger: logger,
	}
}
