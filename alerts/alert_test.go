package alerts

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/mihaichiorean/monidog/mocks"
	"github.com/mihaichiorean/monidog/parser"
	"github.com/stretchr/testify/assert"
)

func Test_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ts := time.Now()
	l := mocks.NewMockLog(ctrl)
	l.EXPECT().Timestamp().Return(ts)
	a := NewAlert("test", 1*time.Second, 1)
	ch := make(chan parser.Log)
	assert.NoError(t, a.Start(ch))
	ch <- l
	assert.NoError(t, a.Stop())
}
