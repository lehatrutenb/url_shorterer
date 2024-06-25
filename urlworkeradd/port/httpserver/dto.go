package httpserver

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (s *server) getNextBatchWaiter() error {
	for i := 0; i < 5; i++ {
		bSt, bLen, dPref, err := s.getNextBatchRequest()
		if err == nil {
			s.lg.Info("Got next bath & date", zap.Int("batch start", bSt), zap.Int("batch len", bLen), zap.String("date prefix", dPref))
			s.mu.Lock()
			s.nxtBFreeAmt = bLen
			s.nxtBInd = bSt
			s.nxtdatePref = dPref
			s.waitingForBatch = false
			s.mu.Unlock()
			return nil
		}
		tmr := time.NewTimer(time.Duration(i*100+rand.Intn(500)) * time.Millisecond)
		<-tmr.C
		tmr.Stop()
	}
	s.mu.Lock()
	s.waitingForBatch = false
	s.mu.Unlock()
	return ErrorServerUnableToGetNewBatch
}

func (s server) getNextBatchRequest() (batch_start int, batch_len int, datePref string, err error) {
	resp, err := http.Get("http://" + s.dataSplitterAddr + "/get")
	if err != nil {
		s.lg.Error("Failed to send request to get new batch", zap.Error(err))
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.lg.Error("Failed to read response body from datasplitter", zap.Error(err))
		return 0, 0, "", err
	}
	data := make(map[string]int)
	if err := json.Unmarshal(body, &data); err != nil {
		s.lg.Error("Failed to marshall response body from datasplitter", zap.Error(err))
		return 0, 0, "", err
	}

	return data["batch_start"], data["batch_len"], s.tMan.ConvertIntTimeToS(data["date_pref"]), nil
}
