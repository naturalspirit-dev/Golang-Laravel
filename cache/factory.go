package cache

import (
	"fmt"
	"github.com/qbhy/goal/contracts"
	"github.com/qbhy/goal/logs"
	"github.com/qbhy/goal/utils"
)

type Factory struct {
	config           contracts.Config
	exceptionHandler contracts.ExceptionHandler
	stores           map[string]contracts.CacheStore
	drivers          map[string]contracts.CacheStoreProvider
}

func (this *Factory) getName(names ...string) string {
	var name string
	if len(names) > 0 {
		name = names[0]
	} else {
		name = this.config.GetString("cache.default")
	}

	return utils.StringOr(name, "default")
}

func (this Factory) getConfig(name string) contracts.Fields {
	return this.config.GetFields(
		utils.IfString(name == "default", "cache", fmt.Sprintf("cache.stores.%s", name)),
	)
}

func (this *Factory) Store(names ...string) contracts.CacheStore {
	name := this.getName(names...)
	if cacheStore, existsStore := this.stores[name]; existsStore {
		return cacheStore
	}

	this.stores[name] = this.get(name)

	return this.stores[name]
}

func (this *Factory) Extend(driver string, cacheStoreProvider contracts.CacheStoreProvider) {
	this.drivers[driver] = cacheStoreProvider
}

func (this *Factory) get(name string) contracts.CacheStore {
	config := this.getConfig(name)
	drive := utils.GetStringField(config, "driver", "redis")
	driveProvider, existsProvider := this.drivers[drive]
	if !existsProvider {
		logs.WithFields(nil).Fatal(fmt.Sprintf("不支持的缓存驱动：%s", drive))
	}
	return driveProvider(config)
}
