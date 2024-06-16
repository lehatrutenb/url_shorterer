package timemanager

import (
	"strconv"
	"time"
)

type TimeManager interface {
	// expected result as any 2 byte string
	GetCurTime() string
}

type TmImpl struct{}

func (tmi TmImpl) GetCurTime() string {
	return strconv.FormatInt(int64(time.Now().UTC().YearDay()), 36)
}
