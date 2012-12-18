package uc

import (
	"net/http"
	"qbox.me/httputil"
	"strconv"
)

type Service struct {
	host, ip string
	Conn     *httputil.Client
}

func New(host, ip string, t http.RoundTripper) (s *Service, err error) {

	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = &Service{host, ip, &httputil.Client{client}}
	return
}

func (s *Service) AntiLeechMode(bucket string, mode int) (code int, err error) {
	param := map[string][]string{
		"bucket": {bucket},
		"mode":   {strconv.Itoa(mode)},
	}
	url := s.ip + "/antiLeechMode"
	return s.Conn.CallWithFormEx(nil, url, s.host, param)
}

func (s *Service) AddAntiLeech(bucket string, mode int, pattern string) (code int, err error) {
	param := map[string][]string{
		"bucket":  {bucket},
		"mode":    {strconv.Itoa(mode)},
		"action":  {"add"},
		"pattern": {pattern},
	}
	url := s.ip + "/referAntiLeech"
	return s.Conn.CallWithFormEx(nil, url, s.host, param)
}

func (s *Service) CleanCache(bucket string) (code int, err error) {
	param := map[string][]string{
		"bucket": {bucket},
	}
	url := s.ip + "/refreshBucket"
	return s.Conn.CallWithFormEx(nil, url, s.host, param)
}

func (s *Service) DelAntiLeech(bucket string, mode int, pattern string) (code int, err error) {
	param := map[string][]string{
		"bucket":  {bucket},
		"mode":    {strconv.Itoa(mode)},
		"action":  {"del"},
		"pattern": {pattern},
	}
	url := s.ip + "/referAntiLeech"
	return s.Conn.CallWithFormEx(nil, url, s.host, param)
}
