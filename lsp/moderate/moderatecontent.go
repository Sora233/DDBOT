package moderate

import (
	"errors"
	"github.com/imroc/req"
	"net"
	"net/http"
	"time"
)

type Config struct {
	ApiKey string
	Host   string
}

var cfg *Config

func InitModerate(apiKey string) {
	req.Debug = true
	c := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
		},
	}
	req.SetClient(c)

	cfg = &Config{
		ApiKey: apiKey,
		Host:   "https://api.moderatecontent.com",
	}
}

func Moderate(url string) (*ModerateResponse, error) {
	param := req.Param{
		"key": cfg.ApiKey,
		"url": url,
	}
	r, err := req.Get(cfg.Host+"/moderate", param)
	if err != nil {
		return nil, err
	}
	if r.Response().StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(r.Response().StatusCode))
	}
	icp := new(ModerateResponse)

	err = r.ToJSON(icp)
	if err != nil {
		return nil, err
	}
	return icp, nil
}

func Anime(url string) (*AnimeResponse, error) {
	param := req.Param{
		"key": cfg.ApiKey,
		"url": url,
	}
	r, err := req.Get(cfg.Host+"/anime", param)
	if err != nil {
		return nil, err
	}
	if r.Response().StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(r.Response().StatusCode))
	}
	icp := new(AnimeResponse)

	err = r.ToJSON(icp)
	if err != nil {
		return nil, err
	}
	return icp, nil
}
