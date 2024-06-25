package httpserver

import (
	"net/http"

	"go.uber.org/zap"
)

// better to move logic to dto/app
func (s *server) urlGetReqsHandler(w http.ResponseWriter, r *http.Request) {
	s.lg.Debug("Got new request to get long url", zap.String("url", r.URL.Path))

	if len(r.URL.Path) <= 1 || r.URL.Path[0] != '/' {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	urlS := r.URL.Path[1:]
	for cRepoTry := 0; cRepoTry < 2; cRepoTry++ {
		url, ok, err := s.cRepo.GetURL(urlS)
		if err != nil {
			s.lg.Error("Got error during getURL to cache repo", zap.Error(err), zap.String("url", urlS))
			//w.WriteHeader(http.StatusInternalServerError) what for?
			//return
		}
		if ok {
			w.Header().Set("Location", url.GetLongURL())
			w.WriteHeader(http.StatusMovedPermanently)
			s.lg.Debug("Send response", zap.String("long url", url.GetLongURL()))
			return
		}
		if cRepoTry == 0 { // do if first try
			s.shBlck[getStringHash(urlS)%int64(len(s.shBlck))].Lock()
			defer s.shBlck[getStringHash(urlS)%int64(len(s.shBlck))].Unlock()
		}
	}

	if _, ok := <-s.repoQue; !ok {
		s.lg.Warn("Repo is too busy")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	url, ok, err := s.repo.GetURL(urlS)
	s.repoQue <- struct{}{}
	if err != nil {
		s.lg.Error("Got error during getURL to repo", zap.Error(err), zap.String("url", urlS))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.cRepo.AddURL(url)

	if ok {
		w.Header().Set("Location", url.GetLongURL())
		w.WriteHeader(http.StatusMovedPermanently)
		s.lg.Debug("Send response", zap.String("short url", r.URL.Path), zap.String("long url", url.GetLongURL()))
		return
	}

	w.WriteHeader(http.StatusNotFound)
	s.lg.Debug("Send response not found")
}
