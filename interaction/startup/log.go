package startup

import (
	"github.com/zmsocc/practice/webook/pkg/logger"
)

func InitLog() logger.Logger {
	return logger.NewNopLogger()
}
