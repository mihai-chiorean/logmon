package reporter

import (
	"testing"
	"time"

	"github.com/Songmu/axslogparser"
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
	r := NewReporter(10 * time.Second)
	ts := time.Now()
	for i := 1000; i > 0; i-- {
		offset := i * 20
		offsetMS := time.Duration(offset) * time.Millisecond
		l := axslogparser.Log{
			RequestURI: "/pages/all/create",
			Time:       ts.Add(-offsetMS),
		}
		r.add(&l)
		if i%2 == 0 {
			l := axslogparser.Log{
				RequestURI: "/pages/half/create",
				Time:       ts.Add(-offsetMS),
			}
			r.add(&l)
		}
		if i%4 == 0 {
			l := axslogparser.Log{
				RequestURI: "/pages/quarter/create",
				Time:       ts.Add(-offsetMS),
			}
			r.add(&l)
		}
	}
	section, _ := r.hotSection()
	assert.Equal(t, "/pages/all", section)
}
