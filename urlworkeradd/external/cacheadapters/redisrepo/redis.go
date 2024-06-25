package redisrepo

import (
	"context"
	"sync"
	"time"
	"urlworkeradd/external/urls" // may be it'll be better not to use structures there, but just key value ~

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisRepo struct {
	client *redis.Client
	ctx    context.Context
	lg     *zap.Logger
	mu     sync.Mutex
}

const AddTimeExp time.Duration = time.Minute * 2

func NewRepo(dbAddr string, ctx context.Context, lg *zap.Logger) *RedisRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     dbAddr,
		Password: "",
		DB:       0,
	})
	return &RedisRepo{client: client, ctx: ctx, lg: lg.With(zap.String("app", "redis cache")), mu: sync.Mutex{}}
}

func (rr *RedisRepo) AddURL(u urls.Urls) {
	rr.lg.Debug("Get request to add url", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()))
	rr.client.Set(rr.ctx, u.GetShortURL(), u.GetLongURL(), AddTimeExp)
}

func (rr *RedisRepo) GetURL(sUrl string) (urls.Urls, bool, error) {
	rr.lg.Debug("Get request to get url", zap.String("short url", sUrl))
	resp := rr.client.Get(rr.ctx, sUrl)
	if err := resp.Err(); err != nil && err != redis.Nil {
		rr.lg.Warn("Failed to get url", zap.Error(err), zap.String("short url", sUrl))
		return urls.Urls{}, false, err
	} else if err == redis.Nil {
		return urls.Urls{}, false, nil
	}
	var u urls.Urls
	u.SetShortURL(sUrl)
	u.SetLongURL(resp.Val())
	rr.lg.Debug("Send response", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()), zap.String("val", resp.Val()))
	return u, true, nil
}

func (rr *RedisRepo) CloseRepo() error {
	if err := rr.client.Close(); err != nil {
		rr.lg.Error("Failed to close cache repo", zap.Error(err))
		return err
	}
	return nil
}
