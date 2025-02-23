package health

import (
	"context"
	"fmt"
	"time"
)

type Reporter func(err error)

type Status struct {
	HasError bool
	Checks   map[string]string
	Status   string
}
type WatchDogOption func(wd *watchDog)

func WithWatchDogInterval(d time.Duration) WatchDogOption {
	return func(wd *watchDog) {
		wd.interval = d
	}
}

func Run(ctx context.Context, options ...WatchDogOption) {

	wd := &watchDog{
		interval: time.Second * 5,
	}
	for _, option := range options {
		option(wd)
	}
	wd.Run(ctx)

}

func GetStatus() Status {
	return data.Status
}

type Option func(*healthCheck)

func WithGracePeriod(d time.Duration) Option {
	return func(hc *healthCheck) {
		hc.GracePeriod = d
	}
}

func WithAutoFailAfter(d time.Duration) Option {
	return func(hc *healthCheck) {
		hc.AutoFailAfter = &d
	}
}

func Register(name string, options ...Option) Reporter {
	dataLock.Lock()
	defer dataLock.Unlock()

	if data.Checks == nil {
		data.Checks = make(map[string]healthCheck)
	}

	check := healthCheck{
		Name:        name,
		GracePeriod: time.Nanosecond * 1,
		Error:       fmt.Errorf("not yet checked"),
	}
	for _, option := range options {
		option(&check)
	}
	if prev, ok := data.Checks[name]; ok {
		check.LastCheck = prev.LastCheck
		check.IsRecovering = prev.IsRecovering
		check.Error = prev.Error
	}
	data.Checks[name] = check

	//data.Status = updateStatus(data)

	return func(err error) {
		Report(name, err)
	}
}

func Report(name string, err error) {
	dataLock.Lock()
	defer dataLock.Unlock()

	if data.Checks == nil {
		data.Checks = make(map[string]healthCheck)
	}

	check, ok := data.Checks[name]
	if !ok {
		check = healthCheck{
			Name:          name,
			AutoFailAfter: nil,
		}
	}

	check.LastCheck = time.Now()
	check.IsRecovering = false
	if check.Error != nil && err == nil {
		check.IsRecovering = true
	}
	check.Error = err
	data.Checks[name] = check

	data.Status = updateStatus(data)
}
