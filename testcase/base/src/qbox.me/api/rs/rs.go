package rs

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	. "qbox.me/api"
	"qbox.me/httputil"
)

type Service struct {
	host, ip map[string]string
	Conn     *httputil.Client
}

func New(host, ip map[string]string, t http.RoundTripper) (s *Service, err error) {

	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = &Service{host, ip, &httputil.Client{client}}
	return
}

type PutRet struct {
	Hash string `json:"hash"`
}

type GetRet struct {
	URL      string `json:"url"`
	Hash     string `json:"hash"`
	MimeType string `json:"mimeType"`
	Fsize    int64  `json:"fsize"`
	Expiry   int64  `json:"expires"`
}

type Entry struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
}

func (s *Service) Put(
	entryURI, mimeType string, body io.Reader, bodyLength int64) (ret PutRet, code int, err error) {
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	url := s.ip["io"] + "/rs-put/" + EncodeURI(entryURI) + "/mimeType/" + EncodeURI(mimeType)
	code, err = s.Conn.CallWithEx(&ret, url, s.host["io"], "application/octet-stream", body, (int64)(bodyLength))
	return
}

// 动态获取文件授权后的临时下载链接
func (s *Service) Get(entryURI, base, attName string, expires int) (data GetRet, code int, err error) {
	url := s.ip["rs"] + "/get/" + EncodeURI(entryURI)
	if base != "" {
		url += "/base/" + base
	}
	if attName != "" {
		url += "/attName/" + EncodeURI(attName)
	}
	if expires > 0 {
		url += "/expires/" + strconv.Itoa(expires)
	}
	//code, err = s.Conn.Call(&data, url)
	code, err = s.Conn.CallEx(&data, url, s.host["rs"])
	if code == 200 {
		data.Expiry += time.Now().Unix()
	}
	return
}

// 
func replaceHostWithIP(url, host, ip string) string {
	return strings.Replace(url, host, ip, 1)
}

// Fetch  downloads a file specified the url and then stores it as the fname
// on the disk.
func (s *Service) Fetch(url, saveAs string) error {
	fi, err := os.OpenFile(saveAs, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	defer fi.Close()
	if err != nil {
		return err
	}

	ip := string([]byte(s.ip["io"][7:]))
	url = replaceHostWithIP(url, s.host["io"], ip)
	reader, err := s.Conn.DownloadEx(url, s.host["io"])
	if err != nil {
		return err
	}
	io.Copy(fi, reader)
	return err
}

func (s *Service) Stat(entryURI string) (entry Entry, code int, err error) {
	code, err = s.Conn.Call(&entry, s.host["rs"]+"/stat/"+EncodeURI(entryURI))
	return
}

func (s *Service) Delete(entryURI string) (code int, err error) {
	return s.Conn.Call(nil, "http://"+s.host["rs"]+"/delete/"+EncodeURI(entryURI))
}

func (s *Service) Mkbucket(bucketname string) (code int, err error) {
	return s.Conn.Call(nil, s.host["rs"]+"/mkbucket/"+bucketname)
}

func (s *Service) Drop(entryURI string) (code int, err error) {
	return s.Conn.Call(nil, s.host["rs"]+"/drop/"+entryURI)
}

func (s *Service) Move(entryURISrc, entryURIDest string) (code int, err error) {
	return s.Conn.Call(nil, s.host["rs"]+"/move/"+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
}

func (s *Service) Copy(entryURISrc, entryURIDest string) (code int, err error) {
	return s.Conn.Call(nil, s.host["rs"]+"/copy/"+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
}

func (s *Service) Publish(domain, table string) (code int, err error) {
	return s.Conn.CallEx(nil, s.ip["rs"]+"/publish/"+EncodeURI(domain)+"/from/"+table, s.host["rs"])
}

func (s *Service) Unpublish(domain string) (code int, err error) {
	return s.Conn.Call(nil, s.host["rs"]+"/unpublish/"+EncodeURI(domain))
}

// -------------------Batcher to do -----------------------------------

type BatchRet struct {
	Data  interface{} `json:"data"`
	Code  int         `json:"code"`
	Error string      `json:"error"`
}

type Batcher struct {
	s1  *Service
	op  []string
	ret []BatchRet
}

func (s *Service) NewBatcher() *Batcher {
	return &Batcher{s1: s}
}

func (b *Batcher) operate(entryURI string, method string) {
	b.op = append(b.op, method+EncodeURI(entryURI))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) operate2(entryURISrc, entryURIDest string, method string) {
	b.op = append(b.op, method+EncodeURI(entryURISrc)+"/"+EncodeURI(entryURIDest))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) Stat(entryURI string) {
	b.operate(entryURI, "/stat/")
}

func (b *Batcher) Get(entryURI string) {
	b.operate(entryURI, "/get/")
}

func (b *Batcher) Delete(entryURI string) {
	b.operate(entryURI, "/delete/")
}

func (b *Batcher) Move(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/move/")
}

func (b *Batcher) Copy(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/copy/")
}

func (b *Batcher) Reset() {
	b.op = nil
	b.ret = nil
}

func (b *Batcher) Len() int {
	return len(b.op)
}

func (b *Batcher) Do() (ret []BatchRet, code int, err error) {
	s := b.s1
	code, err = s.Conn.CallWithForm(&b.ret, s.host["rs"]+"/batch", map[string][]string{"op": b.op})
	ret = b.ret
	return
}

func (s Service) Upload(entryURI, localFile, mimeType, customMeta, callbackParam string, upToken string) (ret PutRet, code int, err error) {

	return s.UploadEx(upToken, localFile, entryURI, mimeType, customMeta, callbackParam, -1, -1)
}

func (s Service) UploadEx(upToken string, localFile, entryURI string, mimeType, customMeta, callbackParam string,
	crc int64, rotate int) (ret PutRet, code int, err error) {

	action := "/rs-put/" + httputil.EncodeURI(entryURI)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	action += "/mimeType/" + httputil.EncodeURI(mimeType)
	if customMeta != "" {
		action += "/meta/" + httputil.EncodeURI(customMeta)
	}
	if crc >= 0 {
		action += "/crc32/" + strconv.FormatInt(crc, 10)
	}
	if rotate >= 0 {
		action += "/rotate/" + strconv.FormatInt(int64(rotate), 10)
	}
	//url := "http://up.qbox.me" + "/upload"
	//url := "http://m1.qbox.me" + "/upload"
	//url := "http://m1.qbox.me" + "/upload"
	url := s.ip["up"] + "/upload"
	multiParams := map[string][]string{
		"action": {action},
		"file":   {"@" + localFile},
		"auth":   {upToken},
	}
	if callbackParam != "" {
		multiParams["params"] = []string{callbackParam}
	}

	code, err = s.Conn.CallWithMultipartEx(&ret, url, s.host["up"], multiParams)
	return
}
