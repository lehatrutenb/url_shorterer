package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

func (s *server) batchReqsHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got new request for a batch")

	s.mu.Lock()
	defer s.mu.Unlock()
	var nBInd, nBSz int = s.lBatchInd, s.batchSize
	if len(s.suBatches) != 0 { // try to send used batch if has any
		nBInd = s.suBatches[len(s.suBatches)-1]
		nBSz = s.batchSize - nBInd%s.batchSize
		s.suBatches = s.suBatches[:len(s.suBatches)-1]
	}
	res, err := json.Marshal(map[string]int{"batch_start": nBInd, "batch_len": nBSz, "date_pref": s.datePref})
	if err != nil {
		s.lg.Error("Failed to marshal response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.lBatchInd += s.batchSize
	w.Write(res)
	w.WriteHeader(http.StatusOK)
	s.lg.Debug("Send response", zap.Int("batch_start", nBInd), zap.Int("batch_len", nBSz))
}

func (s *server) backBatchReqsHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got new request with returned batch")

	buf := make([]byte, 100)
	bSz, err := r.Body.Read(buf)

	if err != nil {
		s.lg.Error("Failed to read req body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var lastUIndex int
	// add simple check to be sure that request is correct
	if lastUIndex, err = strconv.Atoi(string(buf[:bSz])); err != nil || lastUIndex < 0 {
		s.lg.Error("Failed to parse body data", zap.String("data", string(buf[:bSz])), zap.Int("parsed as", lastUIndex), zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.suBatches = append(s.suBatches, lastUIndex)
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	s.lg.Debug("Send response")
}
