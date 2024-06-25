package httpserver

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"urlworkeradd/external/adapters/testrepo"
	"urlworkeradd/external/cacheadapters/testcrepo"
	"urlworkeradd/external/urls"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type testSuite struct {
	suite.Suite
	s server
	t *testing.T
}

func (ts *testSuite) SetupSuite() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to init logger")
	}

	cRepo := testcrepo.NewRepo()
	repo := testrepo.NewRepo()
	ts.s = server{lg: lg, cRepo: &cRepo, repo: &repo}
}

func (ts *testSuite) SetupTest() {
	ts.s.repo.(*testrepo.TestRepo).Clear()
	ts.s.cRepo.(*testcrepo.TestCRepo).Clear()
}

func (ts *testSuite) TestGetReqsRepo() {
	tURL := urls.Urls{}
	tURL.SetShortURL("testurl")
	tURL.SetLongURL("abacaba")
	ts.s.repo.AddURL(tURL)
	req := httptest.NewRequest(http.MethodGet, "/"+tURL.GetShortURL(), nil)
	w := httptest.NewRecorder()

	ts.s.urlGetReqsHandler(w, req)

	resp := w.Result()
	redirect, ok := resp.Header["Location"]
	ts.True(ok)
	ts.Equal([]string{"abacaba"}, redirect)
}

func (ts *testSuite) TestGetReqsCRepo() {
	tURL := urls.Urls{}
	tURL.SetShortURL("testurl")
	tURL.SetLongURL("abacaba")
	ts.s.cRepo.AddURL(tURL)
	req := httptest.NewRequest(http.MethodGet, "/"+tURL.GetShortURL(), nil)
	w := httptest.NewRecorder()

	ts.s.urlGetReqsHandler(w, req)

	resp := w.Result()
	redirect, ok := resp.Header["Location"]
	ts.True(ok)
	ts.Equal([]string{"abacaba"}, redirect)
}

func TestServer(t *testing.T) {
	ts := testSuite{t: t}
	suite.Run(t, &ts)
}
