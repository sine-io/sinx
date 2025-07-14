package ports

import "github.com/sine-io/sinx/internal/job/app"

type GrpcServer struct {
	app app.Application
}

func NewGrpcServer(application app.Application) GrpcServer {
	return GrpcServer{
		app: application,
	}
}
