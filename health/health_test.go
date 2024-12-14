package health

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReportWithNoError(t *testing.T) {
	Report("service1", nil)
	assert.NoError(t, data.Checks["service1"].Error)
	assert.Equal(t, "OK", data.Status.Status)
}

func TestReportWithError(t *testing.T) {
	err := fmt.Errorf("some error")
	Report("service2", err)
	assert.Error(t, data.Checks["service2"].Error)
	assert.Equal(t, "ERROR", data.Status.Status)
}

func TestAutoFailWithoutAutoFailAfter(t *testing.T) {
	Report("service3", nil)
	watchDog{}.autoFail()
	assert.NoError(t, data.Checks["service3"].Error)
}

func TestAutoFailWithAutoFailAfter(t *testing.T) {
	duration := 1 * time.Second
	Report("service4", nil)
	data.Checks["service4"] = healthCheck{
		Name:          "service4",
		AutoFailAfter: &duration,
		LastCheck:     time.Now().Add(-2 * time.Second),
	}
	watchDog{}.autoFail()
	assert.Error(t, data.Checks["service4"].Error)
	assert.Equal(t, "ERROR", data.Status.Status)
}

func TestUpdateStatusWithNoErrors(t *testing.T) {
	data.Checks = map[string]healthCheck{
		"service5": {Name: "service5", Error: nil},
	}
	status := updateStatus(data)
	assert.Equal(t, "OK", status.Status)
}

func TestUpdateStatusWithErrors(t *testing.T) {
	data.Checks = map[string]healthCheck{
		"service6": {Name: "service6", Error: fmt.Errorf("some error")},
	}
	status := updateStatus(data)
	assert.Equal(t, "ERROR", status.Status)
}

func TestRegister(t *testing.T) {
	reporter := Register("service9")
	assert.EqualError(t, data.Checks["service9"].Error, "not yet checked")
	reporter(nil)
	assert.NoError(t, data.Checks["service9"].Error)
}

func TestRegisterInitializesChecksMap(t *testing.T) {
	name := "initialCheck"
	autoFailAfter := 5 * time.Second

	Register(name, WithAutoFailAfter(autoFailAfter))
	dataLock.Lock()
	defer dataLock.Unlock()

	assert.NotNil(t, data.Checks)
	assert.Contains(t, data.Checks, name)
}

func TestRegisterReturnsValidReporter(t *testing.T) {
	name := "reporterCheck"
	autoFailAfter := 5 * time.Second

	reporter := Register(name, WithAutoFailAfter(autoFailAfter))
	assert.NotNil(t, reporter)

	err := fmt.Errorf("test error")
	reporter(err)
	dataLock.Lock()
	defer dataLock.Unlock()

	check, exists := data.Checks[name]
	assert.True(t, exists)
	assert.Equal(t, err, check.Error)
}

func TestReportHandlesNilError(t *testing.T) {
	name := "nilErrorCheck"

	Report(name, nil)
	dataLock.Lock()
	defer dataLock.Unlock()

	check, exists := data.Checks[name]
	assert.True(t, exists)
	assert.Nil(t, check.Error)
	assert.WithinDuration(t, time.Now(), check.LastCheck, time.Second)
}

func TestAutoFailDoesNotUpdateRecentChecks(t *testing.T) {
	name := "recentCheck"
	err := fmt.Errorf("test error")
	autoFailAfter := 5 * time.Minute

	Report(name, err)
	dataLock.Lock()
	data.Checks[name] = healthCheck{
		Name:          name,
		Error:         nil,
		AutoFailAfter: &autoFailAfter,
		LastCheck:     time.Now(),
	}
	dataLock.Unlock()

	watchDog{}.autoFail()
	dataLock.Lock()
	defer dataLock.Unlock()

	check, exists := data.Checks[name]
	assert.True(t, exists)
	assert.Nil(t, check.Error)
}

func TestUpdateStatusHandlesNoChecks(t *testing.T) {
	dataLock.Lock()
	defer dataLock.Unlock()

	data.Checks = make(map[string]healthCheck)
	status := updateStatus(data)
	assert.False(t, status.HasError)
	assert.Equal(t, "OK", status.Status)
	assert.Empty(t, status.Checks)
}
