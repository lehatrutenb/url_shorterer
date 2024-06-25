package httpserver

import (
	"bytes"
	"context"
	"datasplitter/external/timemanager"
	"envconfig"
	"errors"
	"metrics"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"urlworkeradd/external/adapters"
	"urlworkeradd/external/cacheadapters"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// subatches - second used indexes of batches that were returned from killed pods
// better to add another struct for batches vars
type server struct {
	lg               *zap.Logger
	ctx              context.Context
	ctxCncl          context.CancelFunc
	eg               *errgroup.Group
	mu               *sync.Mutex
	es               envconfig.EnvStorage
	metr             *metrics.Metrics
	batchIndNow      int
	batchFreeAmt     int
	datePref         string
	nxtBInd          int
	nxtBFreeAmt      int
	nxtdatePref      string
	dataSplitterAddr string
	waitingForBatch  bool
	cRepo            cacheadapters.CacheRepository
	repo             adapters.Repository
	tMan             timemanager.TimeManager
}

var ErrorSignalServerShutDown error = errors.New("server got signal to shutdown")
var ErrorServerShutDown error = errors.New("server got error to shutdown")
var ErrorServerUnableToGetNewBatch error = errors.New("server all requests to get new batch failed")

func newServer(ctx context.Context, cncl context.CancelFunc, eg *errgroup.Group, lg *zap.Logger, mu *sync.Mutex, es envconfig.EnvStorage, metr *metrics.Metrics, dataSplitterAddr string, cRepo cacheadapters.CacheRepository, repo adapters.Repository, tManager timemanager.TimeManager) server {
	return server{lg: lg.With(zap.String("app", "urlworker adder")), ctx: ctx, ctxCncl: cncl, eg: eg, mu: mu, es: es, metr: metr, dataSplitterAddr: dataSplitterAddr, waitingForBatch: false, cRepo: cRepo, repo: repo, tMan: tManager}
}

func (s *server) init(es envconfig.EnvStorage) error {
	if err := s.getNextBatchWaiter(); err != nil {
		s.lg.Error("Failed to get init batch info")
		return err
	}
	return nil
}

func (s server) shutDown(es envconfig.EnvStorage) {
	s.mu.Lock()
	req, err := http.NewRequest(http.MethodPost, "http:"+s.dataSplitterAddr+"/post", bytes.NewReader([]byte(strconv.FormatInt(int64(s.batchIndNow), 10))))
	if err != nil {
		s.lg.Warn("failed to send request in shutdown")
	}
	s.mu.Unlock()
	http.DefaultClient.Do(req) // we don't need any response - loose is loose
}

func RunServer(parCtx context.Context, lg *zap.Logger, addr string, dataSplitterAddr string, es envconfig.EnvStorage, cRepo cacheadapters.CacheRepository, repo adapters.Repository, metr *metrics.Metrics, tManager timemanager.TimeManager) error {
	var httpSrv http.Server
	httpSrv.Addr = addr

	ctx, cncl := context.WithCancel(parCtx)
	eg, ctx := errgroup.WithContext(ctx)
	mu := &sync.Mutex{}
	server := newServer(ctx, cncl, eg, lg, mu, es, metr, dataSplitterAddr, cRepo, repo, tManager)
	if err := server.init(es); err != nil {
		lg.Error("Failed to init server", zap.Error(err))
		return err
	}

	http.HandleFunc("/post", server.urlAddReqsHandler) // move to router.go

	sigQuit := make(chan os.Signal, 2)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	eg.Go(func() error {
		select {
		case s := <-sigQuit:
			lg.Warn("Shutting server down", zap.Any("signal", s))
		case <-server.ctx.Done():
			lg.Warn("Shutting server down")
		case <-parCtx.Done():
			lg.Warn("Shutting server down")
		}

		if err := httpSrv.Shutdown(ctx); err != nil {
			lg.Info("http server shutdown error: ", zap.Error(err))
		}
		server.shutDown(es)

		return ErrorServerShutDown
	})

	eg.Go(func() error {
		lg.Info("Server is ready for conns")
		return httpSrv.ListenAndServe()
	})

	if err := eg.Wait(); err != nil && err != ErrorServerShutDown && err != http.ErrServerClosed {
		server.lg.Fatal("Server stoppend listening unintentionally", zap.Error(err))
		return err
	}
	server.lg.Info("Server shutted down succesfully")
	return nil
}
