//go:build wireinject

package startup

import (
	"github.com/google/wire"
	repository2 "github.com/zmsocc/practice/webook/interactive/repository"
	cache2 "github.com/zmsocc/practice/webook/interactive/repository/cache"
	dao2 "github.com/zmsocc/practice/webook/interactive/repository/dao"
	service2 "github.com/zmsocc/practice/webook/interactive/service"
)

var thirdProvider = wire.NewSet(
	InitRedis,
	InitTestDB,
	InitLog,
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewInteractiveRepository,
	dao2.NewInteractiveDAO,
	cache2.NewRedisInteractiveCache,
)

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service2.NewInteractiveService(nil, nil)
}
