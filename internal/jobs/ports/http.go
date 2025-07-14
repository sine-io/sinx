package ports

import "github.com/sine-io/sinx/internal/job/app"

type HttpServer struct {
	app app.Application
}

func NewHttpServer(application app.Application) HttpServer {
	return HttpServer{
		app: application,
	}
}
