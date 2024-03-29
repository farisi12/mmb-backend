package main

import (
	"log"
	"net/http"

	"github.com/kodekoding/phastos/go/server"
	"github.com/pkg/errors"
	"github.com/unrolled/secure"

	"godem/infrastructure/config"
	"godem/infrastructure/loader"
	"godem/infrastructure/service/_internal/http/api"
	"godem/infrastructure/service/_internal/router/component"
)

func main() {
	// read config file
	cfg, err := config.New()
	if err != nil {
		log.Fatalln("cannot read config: ", errors.Cause(err).Error())
	}
	modules := loader.InitModules(cfg)
	wrapper := loader.InitWrapper(cfg)
	middlewares := loader.InitMiddleware(modules.Auth)

	handler := api.NewHandler(modules, wrapper, middlewares)
	routeHandler := loadHandler(handler)
	cfgServer := cfg.Server
	cfgServer.Handler = routeHandler
	log.Fatalln(server.ServeHTTP(&cfgServer))
}

func loadHandler(this *api.Handler) http.Handler {
	routes := api.NewRoutes(this)
	// register routes
	routes.Register()

	this.HttpHandler = component.InitHandler(routes.GetHandler())

	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
	})

	this.HttpHandler = secureMiddleware.Handler(this.HttpHandler)

	this.HttpHandler = component.WrapNotif(this.HttpHandler, this.Wrapper().Notif)
	this.HttpHandler = component.WrapApp(this.HttpHandler, this.Wrapper().Apps)
	return this.HttpHandler
}
