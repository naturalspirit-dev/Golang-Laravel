package main

import (
	"github.com/goal-web/application"
	"github.com/goal-web/auth"
	"github.com/goal-web/bloomfilter"
	"github.com/goal-web/cache"
	"github.com/goal-web/config"
	"github.com/goal-web/console/scheduling"
	"github.com/goal-web/contracts"
	"github.com/goal-web/database"
	"github.com/goal-web/email"
	"github.com/goal-web/encryption"
	"github.com/goal-web/events"
	"github.com/goal-web/filesystem"
	"github.com/goal-web/goal/app/console"
	"github.com/goal-web/goal/app/exceptions"
	"github.com/goal-web/goal/app/listeners"
	config2 "github.com/goal-web/goal/config"
	"github.com/goal-web/hashing"
	"github.com/goal-web/http/sse"
	"github.com/goal-web/queue"
	"github.com/goal-web/ratelimiter"
	"github.com/goal-web/redis"
	"github.com/goal-web/serialization"
	"github.com/goal-web/session"
	"github.com/goal-web/supports/logs"
	"github.com/goal-web/websocket"
	"github.com/golang-module/carbon/v2"
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
		config.NewService(config.NewDotEnv(config.File("")), config2.GetConfigProviders()),
		hashing.NewService(),
		encryption.NewService(),
		filesystem.NewService(),
		serialization.NewService(),
		events.NewService(),
		redis.NewService(),
		cache.NewService(),
		bloomfilter.NewService(),
		auth.NewService(),
		ratelimiter.NewService(),
		console.NewService(),
		database.NewService(),
		queue.NewService(false),
		email.NewService(),
		session.NewService(),
		sse.NewService(),
		websocket.NewService(),
		scheduling.NewService(),
	)

	app.Call(func(config contracts.Config, dispatcher contracts.EventDispatcher) {
		appConfig := config.Get("app").(app.Config)
		carbon.SetLocale(appConfig.Locale)
		carbon.SetTimezone(appConfig.Timezone)

		dispatcher.Register("QUERY_EXECUTED", listeners.DebugQuery{})
	})
	if errors := app.Start(); len(errors) > 0 {
		logs.WithField("errors", errors).Fatal("goal 异常!")
	} else {
		logs.Default().Info("goal 已关闭")
	}
}
