package adapters

import "urlworkeradd/external/urls"

// to use urls there is bad idea
type Repository interface {
	GetURL(sURL string) (urls.Urls, bool, error)
	AddURL(u urls.Urls) error
	CloseRepo() error
}
