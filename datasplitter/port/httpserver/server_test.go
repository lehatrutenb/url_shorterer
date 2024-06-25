package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"envconfig"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"datasplitter/external/timemanager"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

type testSuite struct {
	suite.Suite
	s server
	t *testing.T
}

type tmImplTest struct{}

func (tmi tmImplTest) GetCurTime() string {
	return "ab"
}

func (tmi tmImplTest) GetCurTimeInt() int { // a is eq to 36 * 10 ; b to 11
	return 36*10 + 11
}

func (tmi tmImplTest) ConvertIntTimeToS(t int) string {
	return "ab"
}

func (ts *testSuite) SetupSuite() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to init logger")
	}

	eg, ctx := errgroup.WithContext(context.Background())
	ts.s = newServer(10, tmImplTest{}, ctx, eg, lg, &sync.Mutex{})
}

func (ts *testSuite) SetupTest() {
	ts.s.lBatchInd = 0
	ts.s.suBatches = make([]int, 0)
}

func (ts *testSuite) TestGetReqs() {
	for i := 0; i < 1e4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/get", nil)
		w := httptest.NewRecorder()

		ts.s.batchReqsHandler(w, req)

		resp := w.Result()
		buf := make([]byte, 200)
		bSz, err1 := resp.Body.Read(buf)
		data := make(map[string]int)
		err2 := json.Unmarshal(buf[:bSz], &data)
		ts.NoError(err1)
		ts.NoError(err2)
		ts.Equal(http.StatusOK, resp.StatusCode)
		ts.Equal(ts.s.batchSize*i, data["batch_start"])
		ts.Equal(ts.s.batchSize, data["batch_len"])
	}
}

func (ts *testSuite) TestConcurrentGetReqs() {
	mu := &sync.Mutex{}
	var resps []int = make([]int, 0)
	wg := &sync.WaitGroup{}
	for i := 0; i < 1e4; i++ {
		wg.Add(1)
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/get", nil)
			w := httptest.NewRecorder()

			ts.s.batchReqsHandler(w, req)

			resp := w.Result()
			buf := make([]byte, 200)
			bSz, err1 := resp.Body.Read(buf)
			data := make(map[string]int)
			err2 := json.Unmarshal(buf[:bSz], &data)
			mu.Lock()
			resps = append(resps, data["batch_start"])
			mu.Unlock()
			ts.NoError(err1)
			ts.NoError(err2)
			wg.Done()
		}()
	}

	wg.Wait()
	slices.SortFunc(resps, func(x, y int) int {
		return x - y
	})
	for i := 0; i < len(resps); i++ {
		ts.Equal(i*ts.s.batchSize, resps[i])
	}
}

func (ts *testSuite) TestPostAndGetReqs() {
	{ // first get to return back
		req := httptest.NewRequest(http.MethodGet, "/get", nil)
		w := httptest.NewRecorder()

		ts.s.batchReqsHandler(w, req)
	} // already must be tested

	{ // post
		req := httptest.NewRequest(http.MethodPost, "/post", bytes.NewReader([]byte("1")))
		w := httptest.NewRecorder()

		ts.s.backBatchReqsHandler(w, req)
		resp := w.Result()
		ts.Equal(http.StatusOK, resp.StatusCode)
	}

	{ // get
		req := httptest.NewRequest(http.MethodGet, "/get", nil)
		w := httptest.NewRecorder()

		ts.s.batchReqsHandler(w, req)

		resp := w.Result()
		buf := make([]byte, 100)
		bSz, err1 := resp.Body.Read(buf)
		data := make(map[string]int)
		err2 := json.Unmarshal(buf[:bSz], &data)
		ts.NoError(err1)
		ts.NoError(err2)
		ts.Equal(http.StatusOK, resp.StatusCode)
		ts.Equal(1, data["batch_start"])
	}
}

func (ts *testSuite) TestBadPostReqs() {
	reqs := make([]*http.Request, 3)
	reqs[0] = httptest.NewRequest(http.MethodPost, "/post", bytes.NewReader([]byte("-1")))
	reqs[1] = httptest.NewRequest(http.MethodPost, "/post", bytes.NewReader([]byte("abacaba")))
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()

		ts.s.backBatchReqsHandler(w, reqs[i])
		resp := w.Result()
		ts.Equal(http.StatusBadRequest, resp.StatusCode)
	}
}

func TestServer(t *testing.T) {
	ts := testSuite{s: server{}, t: t}
	suite.Run(t, &ts)
}

func doGetRequest(t *testing.T, url string) (batchStart int, batchLen int) {
	resp, err := http.Get("http://localhost:9099/get")
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err1 := io.ReadAll(resp.Body)
	data := make(map[string]int)
	err2 := json.Unmarshal(body, &data)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	return data["batch_start"], data["batch_len"]
}

// care localhost:9099 is tried to be used
func TestShutDown(t *testing.T) {
	es, err := envconfig.NewEnvClientStorage()
	assert.NoError(t, err)
	ctx, cncl := context.WithCancel(context.Background())
	sURL := "localhost:9099"
	server{}.shutDown(es) // to set default "last_batch_index"
	go RunServer(ctx, sURL, 10, timemanager.TmImpl{}, es)

	time.Sleep(100 * time.Millisecond) // better to use pings but now sleep is fine
	doGetRequest(t, sURL)
	cncl()
	time.Sleep(100 * time.Millisecond)

	var s server
	err = s.init(es)
	assert.NoError(t, err)
	assert.Equal(t, 10, s.lBatchInd)
}
