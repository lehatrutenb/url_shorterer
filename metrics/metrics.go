package metrics

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Metrics struct {
	RepoTime  prometheus.Gauge
	CrepoTime prometheus.Gauge
	OtherTime prometheus.Gauge
}

func newMetrics(reg prometheus.Registerer) (*Metrics, error) {
	m := &Metrics{
		prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "repo_duration_sum",
				Help: "",
			},
		),
		prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "crepo_duration_sum",
				Help: "",
			},
		),
		prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "other_duration_sum",
				Help: "",
			},
		),
	}

	err1 := reg.Register(m.RepoTime)
	err2 := reg.Register(m.CrepoTime)
	err3 := reg.Register(m.OtherTime)

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

var ErrorSignalServerShutDown error = errors.New("server got signal to shutdown")
var ErrorServerShutDown error = errors.New("server got error to shutdown")

func ShutDownServer(eg *errgroup.Group, lg *zap.Logger) error {
	lg = lg.With(zap.String("app", "prometheus"))
	if err := eg.Wait(); err != nil && err != ErrorServerShutDown && err != http.ErrServerClosed {
		lg.Fatal("Server stopped listening unintentionally", zap.Error(err))
		return err
	}
	lg.Info("Server shutted down succesfully")
	return nil
}

func RunServer(ctx context.Context, eg *errgroup.Group, lg *zap.Logger, addr string) (*Metrics, error) {
	lg = lg.With(zap.String("app", "prometheus"))
	var httpSrv http.Server
	httpSrv.Addr = addr

	reg := prometheus.NewRegistry()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	sigQuit := make(chan os.Signal, 2)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	eg.Go(func() error {
		select {
		case s := <-sigQuit:
			lg.Warn("Shutting server down", zap.Any("signal", s))
		case <-ctx.Done():
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

	return newMetrics(reg)
}
