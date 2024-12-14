package health

import (
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var (
	dataLock = sync.Mutex{}
	data     = healthData{}
)

type healthData struct {
	Checks map[string]healthCheck
	Status Status
}

type healthCheck struct {
	Name          string
	Error         error
	IsRecovering  bool
	AutoFailAfter *time.Duration
	GracePeriod   time.Duration
	LastCheck     time.Time
}

func updateStatus(db healthData) Status {

	hadError := db.Status.HasError

	status := Status{
		HasError: false,
		Checks:   make(map[string]string),
		Status:   "OK",
	}
	for _, check := range db.Checks {
		if check.Error != nil {
			status.HasError = true
			status.Status = "ERROR"
			status.Checks[check.Name] = check.Error.Error()
			continue
		} else if check.IsRecovering {
			status.HasError = true
			status.Status = "RECOVERING"
			status.Checks[check.Name] = "recovering..."
			continue
		}
		status.Checks[check.Name] = "OK"
	}

	// recover from error
	if hadError && !status.HasError {
		log.Info().Msg("health recovered")
	}

	return status
}
