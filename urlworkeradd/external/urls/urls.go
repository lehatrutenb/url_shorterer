package urls

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Urls struct {
	urlShort string
	urlLong  string
}

func (u Urls) GetShortURL() string {
	return u.urlShort
}

func (u Urls) GetLongURL() string {
	return u.urlLong
}

func (u *Urls) SetShortURL(new string) {
	u.urlShort = new
}

func (u *Urls) SetLongURL(new string) {
	u.urlLong = new
}

type urlst struct {
	UrlShort string
	UrlLong  string
}

func convertFromUrlst(ut urlst) Urls {
	var u Urls
	u.SetLongURL(ut.UrlLong)
	u.SetShortURL(ut.UrlShort)
	return u
}

func convertToUrlst(u Urls) urlst {
	return urlst{u.GetShortURL(), u.GetLongURL()}
}

func DecodeUrlFromBytes(b []byte) (Urls, error) {
	var buf *bytes.Buffer = bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)

	var u urlst
	if err := enc.Decode(&u); err != nil {
		return Urls{}, err
	}
	return convertFromUrlst(u), nil
}

func EncodeUrlToBytes(url Urls) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(convertToUrlst(url)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (u Urls) BeautifulPrint() string {
	return fmt.Sprintf("%s:%s\n", u.urlShort, u.urlLong)
}
