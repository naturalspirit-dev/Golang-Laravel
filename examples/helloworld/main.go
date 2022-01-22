package main

import (
	"github.com/goal-web/application"
	"github.com/goal-web/cache"
	"github.com/goal-web/config"
	"github.com/goal-web/console"
	"github.com/goal-web/contracts"
	"github.com/goal-web/encryption"
	"github.com/goal-web/events"
	"github.com/goal-web/filesystem"
	"github.com/goal-web/hashing"
	"github.com/goal-web/redis"
	"github.com/qbhy/goal/auth"
	"github.com/qbhy/goal/database"
	console2 "github.com/qbhy/goal/examples/helloworld/app/console"
	"github.com/qbhy/goal/examples/helloworld/app/exceptions"
	"github.com/qbhy/goal/examples/helloworld/app/providers"
	config2 "github.com/qbhy/goal/examples/helloworld/config"
	"github.com/qbhy/goal/examples/helloworld/routes"
	"github.com/qbhy/goal/http"
	"github.com/qbhy/goal/session"
	"github.com/qbhy/goal/signal"
	"os"
)

func main() {
	app := application.Singleton()
	path, _ := os.Getwd()
	app.Instance("path", path)

	// 设置异常处理器
	app.Singleton("exceptions.handler", func() contracts.ExceptionHandler {
		return exceptions.NewHandler()
	})

	app.RegisterServices(
		&config.ServiceProvider{
			Env:             os.Getenv("env"),
			Paths:           []string{path},
			Sep:             "=",
			ConfigProviders: config2.Configs(),
		},
		hashing.ServiceProvider{},
		encryption.ServiceProvider{},
		filesystem.ServiceProvider{},
		events.ServiceProvider{},
		redis.ServiceProvider{},
		cache.ServiceProvider{},
		&signal.ServiceProvider{},
		&session.ServiceProvider{},
		auth.ServiceProvider{},
		&database.ServiceProvider{},
		&http.ServiceProvider{RouteCollectors: []interface{}{
			func(router contracts.Router) {
				router.Static("/", "public")
			},
			// 路由收集器
			routes.ApiRoutes,
		}},
		&console.ServiceProvider{
			ConsoleProvider: console2.NewKernel,
		},
		providers.AppServiceProvider{},
	)

	app.Call(func(console3 contracts.Console, input contracts.ConsoleInput) {
		console3.Run(input)
	})
}
