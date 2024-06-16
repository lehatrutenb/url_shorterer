package httpserver

import (
	"context"
	timemanager "datasplitter/internal/time_manager"
	"envconfig"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// subatches - second used indexes of batches that were returned from killed pods
type server struct {
	suBatches []int
	lBatchInd int
	batchSize int
	tManager  timemanager.TimeManager
	lg        *zap.Logger
	ctx       context.Context
	eg        *errgroup.Group
	mu        *sync.Mutex
}

var ErrorSignalServerShutDown error = errors.New("server got signal to shutdown")
var ErrorServerShutDown error = errors.New("server got error to shutdown")

const timeCheckerInterval time.Duration = time.Hour

func newServer(batchSz int, tm timemanager.TimeManager, ctx context.Context, eg *errgroup.Group, lg *zap.Logger, mu *sync.Mutex) server {
	return server{suBatches: make([]int, 0), lBatchInd: 0, batchSize: batchSz, tManager: tm, lg: lg.With(zap.String("app", "datasplitter")), ctx: ctx, eg: eg, mu: mu}
}

func (s *server) init(es envconfig.EnvStorage) error {
	lbis, err := es.EnvGetVal("saved_vals", "last_batch_index")
	if err != nil {
		return err
	}
	lbi, err := strconv.Atoi(lbis)
	if err != nil {
		return err
	}
	s.lBatchInd = lbi
	return err
}

func (s server) shutDown(es envconfig.EnvStorage) error {
	return es.EnvUpdateVal("saved_vals", "last_batch_index", strconv.FormatInt(int64(s.lBatchInd), 10))
}

func (s *server) checkTimeChanging() {
	tckr := time.NewTicker(timeCheckerInterval)
	defer tckr.Stop()
	lastTime := s.tManager.GetCurTime()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-tckr.C:
			if lastTime != s.tManager.GetCurTime() {
				s.lg.Info("Updated date prefix")
				s.mu.Lock()
				s.lBatchInd = 0
				s.suBatches = make([]int, 0)
				s.mu.Unlock()
			}
		}
	}
}

// better to use options there as want base tm, and time for check timechanging
func RunServer(parCtx context.Context, addr string, batchSz int, tm timemanager.TimeManager, es envconfig.EnvStorage) {
	var httpSrv http.Server
	httpSrv.Addr = addr
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to init logger")
	}

	eg, ctx := errgroup.WithContext(parCtx)
	mu := &sync.Mutex{}
	server := newServer(batchSz, tm, ctx, eg, lg, mu)
	if err := server.init(es); err != nil {
		lg.Error("Failed to init datasplitter server, set defaults", zap.Error(err))
	}

	go server.checkTimeChanging()
	http.HandleFunc("/get", server.batchReqsHandler)
	http.HandleFunc("/post", server.backBatchReqsHandler)

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
		if err := server.shutDown(es); err != nil {
			lg.Error("server failed to save env during shutdown")
		}

		return ErrorServerShutDown
	})

	eg.Go(func() error {
		lg.Info("Server is ready for conns")
		return httpSrv.ListenAndServe()
	})

	if err := eg.Wait(); err != nil && err != ErrorServerShutDown && err != http.ErrServerClosed {
		server.lg.Fatal("Server stoppend listening unintentionally", zap.Error(err))
		return
	}
	server.lg.Info("Server shutted down succesfully")
}
