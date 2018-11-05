package alerts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewAlert(t *testing.T) {
	a := NewAlert("test alert", 1*time.Second, 10)
	assert.NotNil(t, a)
}

func Test_Start_NoLogs(t *testing.T) {
	a := NewAlert("test alert", 1*time.Second, 5)
	assert.NotNil(t, a)
	a.Start()
	a.Stop()
}

func Test_Start_multistart(t *testing.T) {
	a := NewAlert("test alert", 1*time.Second, 5)
	assert.NotNil(t, a)
	err := a.Start()
	assert.NoError(t, err)
	assert.Error(t, a.Start())
	a.Stop()
	assert.NoError(t, a.Start())
	a.Stop()
}
