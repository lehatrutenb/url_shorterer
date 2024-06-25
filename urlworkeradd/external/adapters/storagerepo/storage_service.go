package storagerepo

import (
	"context"
	"storage/external/producer"
	"urlworkeradd/external/adapters/postgresrepo"
	"urlworkeradd/external/urls"

	"go.uber.org/zap"
)

type StorageRepo struct {
	pr       *postgresrepo.PostgresRepo
	kAddr    string
	producer *producer.Producer
	lg       *zap.Logger
}

func NewRepo(ctx context.Context, dbAddr string, kAddr string, kTopic string, lg *zap.Logger) *StorageRepo {
	lg = lg.With(zap.String("app", "storage repo"))
	producer, err := producer.NewProducer(ctx, kAddr, lg, kTopic)
	if err != nil {
		lg.Fatal("Failed to connect to storage", zap.Error(err))
	}
	return &StorageRepo{pr: postgresrepo.NewRepo(dbAddr, ctx, lg), kAddr: kAddr, producer: producer, lg: lg}
}

func (sr *StorageRepo) AddURL(u urls.Urls) error {
	sr.lg.Debug("Add url", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()))
	buf, err := urls.EncodeUrlToBytes(u)
	if err != nil {
		sr.lg.Error("Failed to encode urls to bytes", zap.Error(err))
		return err
	}
	sr.producer.WriteMessage(buf)
	return nil
}

func (sr *StorageRepo) GetURL(sURL string) (urls.Urls, bool, error) {
	return sr.pr.GetURL(sURL)
}

func (pr *StorageRepo) CloseRepo() error {
	return nil
}
