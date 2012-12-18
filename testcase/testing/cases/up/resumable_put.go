package up

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"qbox.us/cc/config"
	"qbox.us/log"
	"qbox.me/auth/digest"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/up"
	"qbox.me/api/util"
	"time"
)

type UpResuPut struct {
	Name   string `json:name`
	Bucket string `json:"bucket"`

	Key           string `json:"key"`
	DataFile      string `json:"data_file"`
	DataSha1      string `json:"data_sha1"`
	PutRetryTimes int    `json:"put_retry_times"`
	ExpiresTime   int    `json:"expires_time"`

	ChunkSize int  `json:"chunk_size"`
	BlockBits uint `json:"block_bits"`

	Url      string
	EntryURI string

	Env      api.Env
}

func (self *UpResuPut) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return err
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return err
	}
	self.DataFile = filepath.Join(path, self.DataFile)
	return
}

func (self *UpResuPut) NewRS() (*rs.Service, error) {
	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	return rs.New(self.Env.Hosts, self.Env.Ips, dt)
}

func (self *UpResuPut) doTestPut() (msg string, err error) {

	DataFile := self.DataFile
	entry := self.Bucket + ":" + self.Key
	self.EntryURI = entry
	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	host := self.Env.Hosts["up"]
	ip := self.Env.Ips["up"]
	upservice, _ := up.NewService(host, ip, self.BlockBits, self.ChunkSize, self.PutRetryTimes, dt, 1, 1)
	log.Info(upservice)
	
	f, err := os.Open(DataFile)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	blockCnt := upservice.BlockCount(fi.Size())

	var (
		checksums []string           = make([]string, blockCnt)
		progs     []up.BlockProgress = make([]up.BlockProgress, blockCnt)
		ret       up.PutRet
		code      int
	)
	begin := time.Now()
	code, err = upservice.Put(f, fi.Size(), checksums, progs, func(int, string) {}, func(int, *up.BlockProgress) {})

	if err != nil || code != 200 {
		return
	}
	code, err = upservice.Mkfile(&ret, "/rs-mkfile/", entry, fi.Size(), "", "", checksums)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestPut", begin, end, duration)
	if err != nil || code != 200 {
		return
	}
	return
}

func (self *UpResuPut) doTestRSGet() (msg string, err error) {
	var ret rs.GetRet

	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	rsservice, err := rs.New(self.Env.Hosts, self.Env.Ips, dt)
	if err != nil {
		return
	}
	begin := time.Now()
	ret, code, err := rsservice.Get(self.EntryURI, "", "", 3600)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestRsGet", begin, end, duration)

	if err != nil || code != 200 {
		return
	}
	self.Url = ret.URL
	return
}

func (self *UpResuPut) doTestDownload() (msg string, err error) {
	h := sha1.New()
	begin := time.Now()
	var req *http.Request
	if req, err = http.NewRequest("GET", self.Url, nil); err != nil {
		return
	}
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if _, err = io.Copy(h, resp.Body); err != nil {
		return
	}
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestDownload", begin, end, duration)

	hash := hex.EncodeToString(h.Sum(nil))
	if hash != self.DataSha1 {
		err = errors.New("check shal failed!")
		return
	}
	return
}

func (self *UpResuPut) Test() (msg string, err error) {
	logMsg := func(s string, e error) string {
		msg := ""
		if err == nil {
			msg = fmt.Sprintln(s, " ok")
		} else {
			msg = fmt.Sprintln(s, err)
		}
		return msg
	}

	msg1 := ""
	msg1, err = self.doTestPut()
	msg += logMsg(msg1, err)

	msg1, err = self.doTestRSGet()
	msg += logMsg(msg1, err)

	msg1, err = self.doTestDownload()
	msg += logMsg(msg1, err)

	return
}
