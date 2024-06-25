package testrepo

import "urlworkeradd/external/urls"

type TestRepo struct {
	st map[string]string
}

func NewRepo() TestRepo {
	return TestRepo{make(map[string]string)}
}

func (tr *TestRepo) AddURL(u urls.Urls) error {
	tr.st[u.GetShortURL()] = u.GetLongURL()
	return nil
}

func (tr TestRepo) GetURL(sUrl string) (urls.Urls, bool, error) {
	lUrl, ok := tr.st[sUrl]
	if !ok {
		return urls.Urls{}, false, nil
	}
	url := urls.Urls{}
	url.SetShortURL(sUrl)
	url.SetLongURL(lUrl)
	return url, true, nil
}

func (tr TestRepo) CloseRepo() error {
	return nil
}

func (tr *TestRepo) Clear() {
	tr.st = make(map[string]string)
}
