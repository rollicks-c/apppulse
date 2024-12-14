package apppulse

import (
	"context"
	"github.com/rollicks-c/apppulse/health"
)

func Run(ctx context.Context, options ...health.WatchDogOption) {
	health.Run(ctx, options...)
}

func GetStatus() health.Status {
	return health.GetStatus()
}

func Register(name string, options ...health.Option) health.Reporter {
	return health.Register(name, options...)
}

func Report(name string, err error) {
	health.Report(name, err)
}
