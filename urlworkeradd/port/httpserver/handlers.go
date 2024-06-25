package httpserver

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"urlworkeradd/external/urls"

	"go.uber.org/zap"
)

// better to move logic to dto/app
func (s *server) urlAddReqsHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got new request to short url")

	urlLb, err := io.ReadAll(r.Body)
	if err != nil {
		s.lg.Warn("failed to read request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	urlL := string(urlLb)

	timer := time.Now().UnixMilli()

	var dp string // date pref
	s.mu.Lock()

	if s.batchFreeAmt == 0 {
		if s.nxtBInd != -1 {
			s.batchIndNow = s.nxtBInd
			s.batchFreeAmt = s.nxtBFreeAmt
			s.datePref = s.nxtdatePref
			s.nxtBInd = -1
		} else {
			s.lg.Error("No urls to send")
			w.WriteHeader(http.StatusInternalServerError)
			s.ctxCncl()
			s.mu.Unlock()
			return
		}
	}
	nowURLIndex := s.batchIndNow
	s.batchIndNow++
	s.batchFreeAmt--
	dp = s.datePref
	if s.batchFreeAmt < int(1e4) && !s.waitingForBatch && s.nxtBInd == -1 {
		s.waitingForBatch = true
		s.eg.Go(s.getNextBatchWaiter)
	}
	s.mu.Unlock()

	now := time.Now().UnixMilli()
	s.metr.OtherTime.Set(float64(now - timer))
	timer = now

	var u urls.Urls
	u.SetLongURL(urlL)
	u.SetShortURL(dp + strconv.FormatInt(int64(nowURLIndex), 36))
	s.cRepo.AddURL(u)

	now = time.Now().UnixMilli()
	s.metr.CrepoTime.Set(float64(now - timer))
	timer = now

	s.repo.AddURL(u)

	now = time.Now().UnixMilli()
	s.metr.RepoTime.Set(float64(now - timer))

	w.Write([]byte(u.GetShortURL()))
	w.WriteHeader(http.StatusOK)

	s.lg.Debug("Send response", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()))
}
