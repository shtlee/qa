package up

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
	"net/http"
	"math/rand"
	"path/filepath"
	"qbox.us/cc/config"
	"qbox.me/auth/digest"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/up"
	"qbox.me/api/up2"
	"qbox.me/api/util"
	"qbox.us/errors"
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
	
	Rscli *rs.Service
	Upcli up.Service

	Env      api.Env
}

func (self *UpResuPut) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		err = errors.Info(err, "UpResuPut init failed")
		return
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		err = errors.Info(err, "UpEnv init failed")
		return
	}
	self.DataFile = filepath.Join(path, self.DataFile)
	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	host := self.Env.Hosts["up"]
	ip := self.Env.Ips["up"]
	self.Upcli, err = up.NewService(host, ip, self.BlockBits, self.ChunkSize, self.PutRetryTimes, dt, 2, 2)
	if err != nil {
		err = errors.Info(err, "Upcli init failed")
		return
	}
	self.Rscli, err = rs.New(self.Env.Hosts, self.Env.Ips, dt)
	if err != nil {
		err = errors.Info(err, "Rscli init failed")
		return
	}
	return
}

func (self *UpResuPut) doTestPut() (msg string, err error) {

	DataFile := self.DataFile
	entry := self.Bucket + ":" + self.Key
	self.EntryURI = entry
	
	f, err := os.Open(DataFile)
	if err != nil {
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	blockCnt := self.Upcli.BlockCount(fi.Size())

	var (
		checksums []string           = make([]string, blockCnt)
		progs     []up.BlockProgress = make([]up.BlockProgress, blockCnt)
		ret       up.PutRet
		code      int
	)
	begin := time.Now()
	code, err = self.Upcli.Put(f, fi.Size(), checksums, progs, func(int, string) {}, func(int, *up.BlockProgress) {})

	if err != nil || code != 200 {
		return
	}
	code, err = self.Upcli.Mkfile(&ret, "/rs-mkfile/", entry, fi.Size(), "", "", checksums)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestPut", begin, end, duration)
	if err != nil || code != 200 {
		return
	}
	return
}

func (self *UpResuPut) doTestGet() (msg string, err error) {

	begin := time.Now()
	entryURI := self.Bucket + ":" + self.Key
	ret, code, err := self.Rscli.Get(entryURI, "", "", 3600)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    " + self.Env.Id + "_" + self.Name + "_doTestGet", begin, end, duration)
	if err != nil || code/100 != 2 {
		if err == nil {
			err = errors.New("Invalid response code")
		}
		err = errors.Info(err, "download failed", entryURI)
		return
	}
	resp, err := http.Get(ret.URL)
	if err != nil {
		err = errors.Info(err, "download failed", entryURI, ret.URL)
		return
	}
	defer resp.Body.Close()
	h := sha1.New()
	io.Copy(h, resp.Body)
	hash := hex.EncodeToString(h.Sum(nil))
	if hash != self.DataSha1 {
		err = errors.Info(errors.New("Invalid data sha1"), self.DataSha1, hash)
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

	msg1, err = self.doTestGet()
	msg += logMsg(msg1, err)

	return
}
