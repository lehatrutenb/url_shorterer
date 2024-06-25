package timemanager

import (
	"strconv"
	"time"
)

type TimeManager interface {
	// expected result as any 2 byte string
	GetCurTime() string
	GetCurTimeInt() int
	ConvertIntTimeToS(int) string
}

type TmImpl struct{}

func (tmi TmImpl) GetCurTime() string {
	return strconv.FormatInt(int64(time.Now().UTC().YearDay()), 36)
}

func (tmi TmImpl) GetCurTimeInt() int {
	return time.Now().UTC().YearDay()
}

func (tmi TmImpl) ConvertIntTimeToS(t int) string {
	return strconv.FormatInt(int64(t), 36)
}
