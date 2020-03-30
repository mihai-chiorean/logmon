package reporter

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/mihaichiorean/monidog/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_parseSection(t *testing.T) {
	s, err := parseSection("/pages/create")
	assert.NoError(t, err)
	assert.Equal(t, "/pages", s)
	s, err = parseSection("/pages/create/lol.php")
	assert.NoError(t, err)
	assert.Equal(t, "/pages/create", s)
	s, err = parseSection("/pages")
	assert.NoError(t, err)
	assert.Equal(t, "/pages", s)
	s, err = parseSection("/")
	assert.NoError(t, err)
	assert.Equal(t, "/", s)
}

func Test_NewReporter(t *testing.T) {
	r := NewReporter(10 * time.Second)
	assert.NotNil(t, r)
	assert.Equal(t, cap(r.buckets), 10)
	assert.Equal(t, 1000*time.Millisecond, r.bucketMS)
}

func Test_hotSection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ts := time.Now()
	r := NewReporter(10 * time.Second)
	for i := 1000; i > 0; i-- {
		offset := i * 20
		offsetMS := time.Duration(offset) * time.Millisecond
		l := mocks.NewMockLog(ctrl)
		l.EXPECT().Timestamp().Return(ts.Add(-offsetMS))
		l.EXPECT().Resource().Return("/pages/all/create")
		r.add(l)
		if i%2 == 0 {
			l := mocks.NewMockLog(ctrl)
			l.EXPECT().Timestamp().Return(ts.Add(-offsetMS))
			l.EXPECT().Resource().Return("/pages/half/create")
			r.add(l)
		}
		if i%4 == 0 {
			l := mocks.NewMockLog(ctrl)
			l.EXPECT().Timestamp().Return(ts.Add(-offsetMS))
			l.EXPECT().Resource().Return("/pages/quarter/create")
			r.add(l)
		}
	}
	section, _ := r.hotSection()
	assert.Equal(t, "/pages/all", section)
}
