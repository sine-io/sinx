package service

import (
	"context"

	"github.com/sine-io/sinx/internal/job/app"
)

func NewApplication(ctx context.Context) app.Application {
	// other logic

	return app.Application{
		Commands: app.Commands{},
		Queries:  app.Queries{},
	}
}
