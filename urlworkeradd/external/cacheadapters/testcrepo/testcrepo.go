package testcrepo

import "urlworkeradd/external/urls"

type TestCRepo struct {
	st map[string]string
}

func NewRepo() TestCRepo {
	return TestCRepo{make(map[string]string)}
}

func (tr *TestCRepo) AddURL(u urls.Urls) {
	tr.st[u.GetShortURL()] = u.GetLongURL()
}

func (tr TestCRepo) GetURL(sUrl string) (urls.Urls, bool, error) {
	lUrl, ok := tr.st[sUrl]
	if !ok {
		return urls.Urls{}, false, nil
	}
	url := urls.Urls{}
	url.SetShortURL(sUrl)
	url.SetLongURL(lUrl)
	return url, true, nil
}

func (tr TestCRepo) CloseRepo() error {
	return nil
}

func (tr *TestCRepo) Clear() {
	tr.st = make(map[string]string)
}
