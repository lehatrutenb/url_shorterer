package httpserver

import (
	"context"
	"datasplitter/external/timemanager"
	"envconfig"
	"errors"
	"net/http"
	"os"
	"os/signal"
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
	lg      *zap.Logger
	ctx     context.Context
	ctxCncl context.CancelFunc
	eg      *errgroup.Group
	mu      *sync.Mutex
	es      envconfig.EnvStorage
	cRepo   cacheadapters.CacheRepository
	repo    adapters.Repository
	tMan    timemanager.TimeManager
	shBlck  []*sync.Mutex
	repoQue chan struct{}
}

var ErrorSignalServerShutDown error = errors.New("server got signal to shutdown")
var ErrorServerShutDown error = errors.New("server got error to shutdown")

const maxRepoWorkers = 30

func newServer(ctx context.Context, cncl context.CancelFunc, eg *errgroup.Group, lg *zap.Logger, mu *sync.Mutex, es envconfig.EnvStorage, cRepo cacheadapters.CacheRepository, repo adapters.Repository, tMan timemanager.TimeManager, shAmount int) server {
	shBlck := make([]*sync.Mutex, shAmount)
	for i := 0; i < shAmount; i++ {
		shBlck[i] = &sync.Mutex{}
	}
	repoQue := make(chan struct{}, maxRepoWorkers)
	for i := 0; i < maxRepoWorkers; i++ {
		repoQue <- struct{}{}
	}
	return server{lg: lg.With(zap.String("app", "urlworker getter")), ctx: ctx, ctxCncl: cncl, eg: eg, mu: mu, es: es, cRepo: cRepo, repo: repo, tMan: tMan, shBlck: shBlck, repoQue: repoQue}
}

func RunServer(parCtx context.Context, lg *zap.Logger, addr string, es envconfig.EnvStorage, cRepo cacheadapters.CacheRepository, repo adapters.Repository, tMan timemanager.TimeManager, shAmount int) error {
	var httpSrv http.Server
	httpSrv.Addr = addr

	ctx, cncl := context.WithCancel(parCtx)
	eg, ctx := errgroup.WithContext(ctx)
	mu := &sync.Mutex{}
	server := newServer(ctx, cncl, eg, lg, mu, es, cRepo, repo, tMan, shAmount)

	http.HandleFunc("/", server.urlGetReqsHandler)

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

		return ErrorServerShutDown
	})

	eg.Go(func() error {
		lg.Info("Server is ready for conns")
		return httpSrv.ListenAndServe()
	})

	if err := eg.Wait(); err != nil && err != ErrorServerShutDown && err != http.ErrServerClosed {
		lg.Fatal("Server stoppend listening unintentionally", zap.Error(err))
		return err
	}
	lg.Info("Server shutted down succesfully")
	return nil
}
