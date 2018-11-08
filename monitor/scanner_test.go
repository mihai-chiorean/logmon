package monitor

import (
	"fmt"
	"io"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/mihaichiorean/monidog/mocks"
	"github.com/mihaichiorean/monidog/parser"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_Watch_fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	scanner, err := Watch(mockSeekReader, nil, 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil provided")
	assert.Nil(t, scanner)
}

func Test_Watch_OK(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p := mocks.NewMockLogParser(mockCtrl)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	fi := mocks.NewMockFileInfo(mockCtrl)
	mockSeekReader.EXPECT().Stat().Return(fi, nil)
	scanner, err := Watch(mockSeekReader, p, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, scanner)
	assert.NoError(t, scanner.Close())
}
func Test_Subscribe(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p := mocks.NewMockLogParser(mockCtrl)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	fi := mocks.NewMockFileInfo(mockCtrl)
	mockSeekReader.EXPECT().Stat().Return(fi, nil)
	scanner, err := Watch(mockSeekReader, p, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, scanner)

	ch := scanner.Subscribe()
	assert.NotNil(t, ch)

	testClose := make(chan error)
	go func() {
		_, ok := <-ch
		if ok {
			testClose <- fmt.Errorf("not closed")
			return
		}
		testClose <- nil
	}()

	assert.NoError(t, scanner.Close())
	err = <-testClose
	assert.NoError(t, err)
}

func Test_hasChanged(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p := mocks.NewMockLogParser(mockCtrl)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	fiBefore := mocks.NewMockFileInfo(mockCtrl)
	fiAfter := mocks.NewMockFileInfo(mockCtrl)
	ls := logScanner{
		SugaredLogger: zap.NewNop().Sugar(),
		interval:      1 * time.Second,
		subscribing:   make(chan chan parser.Log),
		closing:       make(chan chan error),
		parser:        p,
	}

	// same size
	fiBefore.EXPECT().Size().Return(int64(0))
	fiAfter.EXPECT().Size().Return(int64(0))
	mockSeekReader.EXPECT().Stat().Return(fiAfter, nil)
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(0), nil)
	pos, changed, err := ls.hasChanged(mockSeekReader, fiBefore)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pos)
	assert.False(t, changed)

	// size increased
	fiBefore.EXPECT().Size().Return(int64(0)).Times(2)
	fiAfter.EXPECT().Size().Return(int64(64)).Times(2)
	mockSeekReader.EXPECT().Stat().Return(fiAfter, nil)
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(0), nil)
	pos, changed, err = ls.hasChanged(mockSeekReader, fiBefore)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pos)
	assert.True(t, changed)

	// size decrease
	fiBefore.EXPECT().Size().Return(int64(64)).Times(2)
	fiAfter.EXPECT().Size().Return(int64(10)).Times(2)
	mockSeekReader.EXPECT().Stat().Return(fiAfter, nil)
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(1), nil)
	pos, changed, err = ls.hasChanged(mockSeekReader, fiBefore)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), pos)
	assert.True(t, changed)
}

func Test_hasChanged_fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	p := mocks.NewMockLogParser(mockCtrl)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	fiBefore := mocks.NewMockFileInfo(mockCtrl)
	ls := logScanner{
		SugaredLogger: zap.NewNop().Sugar(),
		interval:      1 * time.Second,
		subscribing:   make(chan chan parser.Log),
		closing:       make(chan chan error),
		parser:        p,
	}

	// seek failed
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(0), fmt.Errorf("seek failed"))
	_, changed, err := ls.hasChanged(mockSeekReader, fiBefore)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "seek failed")
	assert.False(t, changed)

	// seek failed
	mockSeekReader.EXPECT().Stat().Return(nil, fmt.Errorf("stat failed"))
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(0), nil)
	_, changed, err = ls.hasChanged(mockSeekReader, fiBefore)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stat failed")
	assert.False(t, changed)
}
func Test_loop(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	line := `127.0.0.1 - lol [06/Nov/2018:14:31:29 -0800] "OPTIONS /pages/subpages/create HTTP/1.0" 201 8582`
	p := mocks.NewMockLogParser(mockCtrl)
	mockSeekReader := mocks.NewMockSeekReader(mockCtrl)
	fiBefore := mocks.NewMockFileInfo(mockCtrl)
	fiAfter := mocks.NewMockFileInfo(mockCtrl)
	// same size
	fiBefore.EXPECT().Size().Return(int64(0)).Times(2)
	fiAfter.EXPECT().Size().Return(int64(len(line))).Times(2)
	call1 := mockSeekReader.EXPECT().Stat().Return(fiBefore, nil)
	call2 := mockSeekReader.EXPECT().Stat().Return(fiAfter, nil)
	call2.After(call1)
	mockSeekReader.EXPECT().Seek(int64(0), io.SeekCurrent).Return(int64(0), nil)
	mockSeekReader.EXPECT().Seek(int64(0), 0)
	p.EXPECT().Parse(gomock.Any()).Return(mocks.NewMockLog(mockCtrl), nil)
	scan1 := mockSeekReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		copy(b, []byte(line))
		return len(b), nil
	})
	scan1.SetArg(0, []byte(line))
	scan2 := mockSeekReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		return 0, io.EOF
	})
	scan2.After(scan1)
	ls, err := Watch(mockSeekReader, p, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, ls)

	ch := ls.Subscribe()
	assert.NotNil(t, ch)

	testClose := make(chan parser.Log)
	go func() {
		v, ok := <-ch
		if ok {
			testClose <- v
			return
		}
		testClose <- nil
	}()

	l := <-testClose
	assert.NoError(t, ls.Close())
	assert.NotNil(t, l)
}
