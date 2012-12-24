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
	"qbox.me/api/up2"
	"qbox.me/api/util"
	"qbox.us/errors"
)

type UpRPut struct {
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
	Up2cli *up2.Service

	Env      api.Env
}

func (self *UpRPut) Init(conf, env, path string) (err error) {

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

	self.Rscli, err = rs.New(self.Env.Hosts, self.Env.Ips, dt)
	if err != nil {
		err = errors.Info(err, "Rscli init failed")
		return
	}
	self.Up2cli, err = up2.New(host, ip, self.BlockBits, self.ChunkSize, self.PutRetryTimes, dt)
	if err != nil {
		err = errors.Info(err, "Up2cli init failed")
		return
	}
	return
}

func (self *UpRPut) doTestRPut() (msg string, err error) {

	f, err := os.Open(self.DataFile)
	if err != nil {
		err = errors.Info(err, "Resumable put failed")
		return
	}
	defer f.Close()
	fi, _ := f.Stat()
	entryURI := self.Bucket + ":" + self.Key
	blockcnt := self.Up2cli.BlockCount(fi.Size())
	progs := make([]up2.BlockputProgress, blockcnt)
	
	chunkNotify := func(idx int, p *up2.BlockputProgress) {
		if rand.Intn(blockcnt)/3 == 0 {
			p1 := *p
			progs[idx] = p1
		}
	}
	blockNotify := func(idx int, p *up2.BlockputProgress) {
	}
	t1 := self.Up2cli.NewRPtask(entryURI, "", "", "", "", f, fi.Size(), nil)
	t1.ChunkNotify = chunkNotify
	t1.BlockNotify = blockNotify

	begin := time.Now()
	for i := 0; i < blockcnt; i++ {
		t1.PutBlock(i)
	}
	t1.Progress = progs
	code, err := t1.Run(10, 10, nil, nil)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("UP    "+self.Env.Id+"_"+self.Name+"_doTestRPut", begin, end, duration)
	if err != nil || code/100 != 2 {
		err = errors.Info(errors.New("Resumable put failed"), entryURI, err, code)
		return
	}
	return
}


func (self *UpRPut) doTestGet() (msg string, err error) {

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


func (self *UpRPut) Test() (msg string, err error) {
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

	msg1, err = self.doTestRPut()
	msg += logMsg(msg1, err)
	msg1, err = self.doTestGet()
	msg += logMsg(msg1, err)

	return
}
