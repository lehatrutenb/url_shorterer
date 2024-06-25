package cacheadapters

import "urlworkeradd/external/urls"

// to use urls there is bad idea
type CacheRepository interface {
	AddURL(u urls.Urls)
	GetURL(sUrl string) (urls.Urls, bool, error)
	CloseRepo() error
}
