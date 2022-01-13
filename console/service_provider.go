package console

import (
	"github.com/golang-module/carbon/v2"
	"github.com/gorhill/cronexpr"
	"github.com/qbhy/goal/application"
	"github.com/qbhy/goal/console/inputs"
	"github.com/qbhy/goal/contracts"
	"github.com/qbhy/goal/exceptions"
	"github.com/qbhy/goal/logs"
	"github.com/qbhy/goal/utils"
	"time"
)

type ConsoleProvider func(application contracts.Application) contracts.Console

type ServiceProvider struct {
	ConsoleProvider ConsoleProvider

	stopChan     chan bool
	serverIdChan chan bool
	app          contracts.Application
	execRecords  map[int]time.Time
}

func (this *ServiceProvider) Register(application contracts.Application) {
	this.serverIdChan = make(chan bool, 1)
	this.app = application
	this.execRecords = make(map[int]time.Time)

	application.Singleton("console", func() contracts.Console {
		console := this.ConsoleProvider(application)
		console.Schedule(console.GetSchedule())
		return console
	})
	application.Singleton("scheduling", func(console contracts.Console) contracts.Schedule {
		return console.GetSchedule()
	})
	application.Singleton("console.input", func() contracts.ConsoleInput {
		return inputs.NewOSArgsInput()
	})
}

func (this *ServiceProvider) runScheduleEvents(events []contracts.ScheduleEvent) {
	if len(events) > 0 {
		// 并发执行所有事件
		now := time.Now()
		for index, event := range events {
			lastExecTime := this.execRecords[index]
			nextTime := carbon.Time2Carbon(cronexpr.MustParse(event.Expression()).Next(lastExecTime))
			nowCarbon := carbon.Time2Carbon(now)
			if nextTime.DiffInSeconds(nowCarbon) == 0 {
				this.execRecords[index] = now
				go (func(event contracts.ScheduleEvent) {
					event.Run(this.app)
				})(event)
			} else if nextTime.Lt(nowCarbon) {
				this.execRecords[index] = now
			}
		}
	}
}

func (this *ServiceProvider) Start() error {
	go this.maintainServerId()
	this.app.Call(func(schedule contracts.Schedule) {
		this.stopChan = utils.SetInterval(1, func() {
			this.runScheduleEvents(schedule.GetEvents())
		}, func() {
			logs.Default().Info("the goal scheduling is closed")
		})
	})
	return nil
}

func (this *ServiceProvider) Stop() {
	this.stopChan <- true
	if this.serverIdChan != nil {
		this.serverIdChan <- true
	}
}

func (this *ServiceProvider) maintainServerId() {
	this.app.Call(func(redis contracts.RedisConnection, config contracts.Config, handler contracts.ExceptionHandler) {
		appConfig := config.Get("app").(application.Config)
		_, err := redis.SAdd("goal.servers", appConfig.ServerId)
		if err != nil {
			handler.Handle(exceptions.WithError(err, contracts.Fields{
				"appConfig": appConfig,
			}))
			return
		}
		this.serverIdChan = utils.SetInterval(1, func() {
			// 维持当前服务心跳
			_, _ = redis.Set("goal.server."+appConfig.ServerId, time.Now().String(), time.Second*5)

			// 挂掉的服务就删掉
			servers, _ := redis.SMembers("goal.servers")
			for _, serverId := range servers {
				if num, _ := redis.Exists("goal.server." + serverId); num == 0 {
					_, _ = redis.SRem("goal.servers", serverId)
				}
			}
		}, func() {
			_, _ = redis.Del("goal.server." + appConfig.ServerId)
			_, _ = redis.SRem("goal.servers", appConfig.ServerId)
		})
	})
}
