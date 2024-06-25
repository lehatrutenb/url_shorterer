package postgresrepo

import (
	"context"
	"time"
	"urlworkeradd/external/urls"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresRepo struct {
	conn *pgxpool.Pool
	lg   *zap.Logger
	ctx  context.Context
}

func NewRepo(dbAddr string, ctx context.Context, lg *zap.Logger) *PostgresRepo {
	//conn, err := pgx.Connect(ctx, dbAddr)
	conn, err := pgxpool.New(ctx, dbAddr)
	if err != nil {
		lg.Fatal("Failed to connect to postgres repo", zap.Error(err))
	}
	return &PostgresRepo{conn: conn, lg: lg.With(zap.String("app", "postgresrepo")), ctx: ctx}
}

const getLongURLQuery = `SELECT longurl FROM urls WHERE shorturl = $1 LIMIT 1`

func (pr *PostgresRepo) GetURL(sURL string) (urls.Urls, bool, error) {
	pr.lg.Debug("Get url request", zap.String("short url", sURL))
	rows, err := pr.conn.Query(pr.ctx, getLongURLQuery, sURL)
	if err != nil {
		pr.lg.Error("Failed to query urls from repo", zap.Error(err), zap.String("Short url", sURL))
		return urls.Urls{}, false, err
	}

	ok := rows.Next()
	if rows.Err() != nil {
		pr.lg.Error("Error occured during go over rows", zap.Error(err))
		return urls.Urls{}, false, err
	}
	if !ok {
		return urls.Urls{}, false, nil
	}

	defer rows.Close()

	var lURL string
	if err := rows.Scan(&lURL); err != nil {
		pr.lg.Error("Failed to scan selected rows", zap.Error(err))
		return urls.Urls{}, false, err
	}
	var url urls.Urls = urls.Urls{}
	url.SetShortURL(sURL)
	url.SetLongURL(lURL)

	return url, true, nil
}

const addURLQuery = `INSERT INTO urls (shorturl, longurl, timestamp) VALUES ($1, $2, $3)`

func (pr *PostgresRepo) AddURL(u urls.Urls) error {
	pr.lg.Debug("Add url", zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()))
	_, err := pr.conn.Exec(pr.ctx, addURLQuery, u.GetShortURL(), u.GetLongURL(), time.Now().UnixMilli())
	if err != nil {
		pr.lg.Error("Failed to add url to repo", zap.Error(err), zap.String("short url", u.GetShortURL()), zap.String("long url", u.GetLongURL()))
		return err
	}
	return nil
}

func (pr *PostgresRepo) CloseRepo() error {
	pr.conn.Close()
	return nil
}
