package pub

import (
	"net/http"
	"qbox.me/httputil"
	"strconv"
	"qbox.us/rpc"
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
		url += "/from/" + rpc.EncodeURI(srcSiteUrl)
	}
	if expires != 0 {
		url += "/expires/" + strconv.Itoa(expires)
	}
	if srcHost != "" {
		url += "/host/" + rpc.EncodeURI(srcHost)
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
	return s.Conn.CallEx(nil, s.host+"/separator/"+bucketName+"/sep/"+rpc.EncodeURI(sep), s.ip)
}

func (s *Service) Style(bucketName string, name string, style string) (code int, err error) {
	url := s.host + "/style/" + bucketName + "/name/" + rpc.EncodeURI(name) + "/style/" + rpc.EncodeURI(style)
	return s.Conn.CallEx(nil, url, s.ip)
}

func (s *Service) Unstyle(bucketName string, name string) (code int, err error) {
	return s.Conn.CallEx(nil, s.host+"/unstyle/"+bucketName+"/name/"+rpc.EncodeURI(name), s.ip)
}
