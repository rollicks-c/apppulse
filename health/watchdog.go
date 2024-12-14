package health

import (
	"context"
	"fmt"
	"time"
)

type watchDog struct {
	interval time.Duration
}

func (wd watchDog) Run(ctx context.Context) {

	// setup timer
	ticker := time.NewTicker(wd.interval)
	defer ticker.Stop()

	for {

		// invoke time-based updates
		wd.autoFail()
		wd.recoverFromError()

		// await next tick
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}

	}
}

func (wd watchDog) autoFail() {

	dataLock.Lock()
	defer dataLock.Unlock()

	now := time.Now()
	for name, check := range data.Checks {
		if check.AutoFailAfter == nil {
			continue
		}
		if check.LastCheck.Add(*check.AutoFailAfter).Before(now) {
			check.Error = fmt.Errorf("no report since %s", check.LastCheck.Format(time.RFC1123))
			data.Checks[name] = check
		}
	}

	data.Status = updateStatus(data)
}

func (wd watchDog) recoverFromError() {

	dataLock.Lock()
	defer dataLock.Unlock()

	now := time.Now()
	for name, check := range data.Checks {
		if !check.IsRecovering {
			continue
		}
		if check.LastCheck.Add(check.GracePeriod).Before(now) {
			check.Error = nil
			check.IsRecovering = false
			data.Checks[name] = check
		}
	}

	data.Status = updateStatus(data)
}
