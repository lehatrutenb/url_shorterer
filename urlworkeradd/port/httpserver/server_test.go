package httpserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"urlworkeradd/external/adapters/testrepo"
	"urlworkeradd/external/cacheadapters/testcrepo"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type testSuite struct {
	suite.Suite
	s server
	t *testing.T
}

func (ts *testSuite) SetupSuite() {
	lg, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to init logger")
	}

	cRepo := testcrepo.NewRepo()
	repo := testrepo.NewRepo()
	ts.s = server{lg: lg, cRepo: &cRepo, repo: &repo, batchIndNow: 0, batchFreeAmt: int(1e6), mu: &sync.Mutex{}}
}

func (ts *testSuite) SetupTest() {
	ts.s.repo.(*testrepo.TestRepo).Clear()
	ts.s.cRepo.(*testcrepo.TestCRepo).Clear()
}

func (ts *testSuite) TestGetReqsRepo() {
	ts.s.batchIndNow = 10
	req := httptest.NewRequest(http.MethodPost, "/post", bytes.NewReader([]byte("testurl")))
	w := httptest.NewRecorder()

	ts.s.urlAddReqsHandler(w, req)

	resp := w.Result()
	buf := make([]byte, 100)
	bSz, err := resp.Body.Read(buf)
	ts.NoError(err)
	sURL := string(buf[:bSz])
	ts.Equal("a", sURL)

	lURL1, ok1, err1 := ts.s.cRepo.GetURL(sURL)
	lURL2, ok2, err2 := ts.s.repo.GetURL(sURL)
	ts.NoError(err1)
	ts.True(ok1)
	ts.Equal("testurl", lURL1.GetLongURL())
	ts.NoError(err2)
	ts.True(ok2)
	ts.Equal("testurl", lURL2.GetLongURL())
}

func TestServer(t *testing.T) {
	ts := testSuite{t: t}
	suite.Run(t, &ts)
}
