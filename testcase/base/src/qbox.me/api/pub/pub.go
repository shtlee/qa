package pub

import (
	"net/http"
	. "qbox.me/api"
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

type BucketInfo struct {
	Source    string            `json:"source" bson:"source"`
	Host      string            `json:"host" bson:"host"`
	Expires   int               `json:"expires" bson:"expires"`
	Protected int               `json:"protected" bson:"protected"`
	Separator string            `json:"separator" bson:"separator"`
	Styles    map[string]string `json:"styles" bson:"styles"`
}

func (s *Service) Image(bucketName string, srcSiteUrls []string, srcHost string, expires int) (code int, err error) {
	url := s.host + "/image/" + bucketName
	for _, srcSiteUrl := range srcSiteUrls {
		url += "/from/" + EncodeURI(srcSiteUrl)
	}
	if expires != 0 {
		url += "/expires/" + strconv.Itoa(expires)
	}
	if srcHost != "" {
		url += "/host/" + EncodeURI(srcHost)
	}
	return s.Conn.CallEx(nil, url, s.ip)
}

func (s *Service) Unimage(bucketName string) (code int, err error) {
	return s.Conn.CallEx(nil, s.host+"/unimage/"+bucketName, s.ip)
}

func (s *Service) Info(bucketName string) (info BucketInfo, code int, err error) {
	code, err = s.Conn.CallEx(&info, s.host+"/info/"+bucketName, s.ip)
	return
}

func (s *Service) AccessMode(bucketName string, mode int) (code int, err error) {
	return s.Conn.CallEx(nil, s.host+"/accessMode/"+bucketName+"/mode/"+strconv.Itoa(mode), s.ip)
}

func (s *Service) Separator(bucketName string, sep string) (code int, err error) {
	return s.Conn.CallEx(nil, s.host+"/separator/"+bucketName+"/sep/"+EncodeURI(sep), s.ip)
}

func (s *Service) Style(bucketName string, name string, style string) (code int, err error) {
	url := s.host + "/style/" + bucketName + "/name/" + EncodeURI(name) + "/style/" + EncodeURI(style)
	return s.Conn.CallEx(nil, url, s.ip)
}

func (s *Service) Unstyle(bucketName string, name string) (code int, err error) {
	return s.Conn.CallEx(nil, s.host+"/unstyle/"+bucketName+"/name/"+EncodeURI(name), s.ip)
}
