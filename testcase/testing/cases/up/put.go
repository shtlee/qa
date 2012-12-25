package up

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"qbox.us/cc/config"
	da "qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/util"
	"time"
)

type PutFile struct {
	Name       string `json:"name"`
	DataFile string `json:"data_file"`
	BucketName string `json:"bucket"`
	Key        string `json:"key"`
	Sha1       string `json:"data_sha1"`

	Conn *rs.Service
	Env  api.Env
}

func (self *PutFile) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return err
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return err
	}
	dt := da.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	self.Conn, err = rs.New(self.Env.Hosts, self.Env.Ips, dt)
	self.DataFile = filepath.Join(path, self.DataFile)
	return
}

// upload the file and get the download url 
func (self *PutFile) doTestPutFile() (url, msg string, err error) {
	entry := self.BucketName + ":" + self.Key
	authPolicy := &uptoken.AuthPolicy{
		Scope:    entry,
		Deadline: 3600,
	}
	authPolicy.Deadline += uint32(time.Now().Unix())
	token := uptoken.MakeAuthTokenString(self.Env.AccessKey, self.Env.SecretKey, authPolicy)

	// in fact, upload should be a part of Up not Rs
	begin := time.Now()
	_, code, err := self.Conn.Upload(entry, self.DataFile, "", "", "", token)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("UP    "+self.Env.Id+"_"+self.Name+"_doTestPutFile", begin, end, duration)
	if err != nil || code != 200 {
		return
	}

	getRet, code, err := self.Conn.Get(entry, "", "", 3600)
	if err != nil || code != 200 {
		return
	}
	url = getRet.URL
	// check shal
	return
}

func (self *PutFile) doTestCheckSha1(url string) (msg string, err error) {
	begin := time.Now()
	netBuf, err := util.DoHttpGet(url)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("RS    "+self.Env.Id+"_"+self.Name+"_doTestIoDownload", begin, end, duration)
	if err != nil {
		return
	}

	h := sha1.New()
	_, err = io.Copy(h, netBuf)
	if err != nil {
		return
	}
	sha1Code := hex.EncodeToString(h.Sum(nil))
	if self.Sha1 != sha1Code {
		err = errors.New("check sha1 failed!")
	}
	return
}

func (self *PutFile) Test() (msg string, err error) {
	msg1 := ""
	url, msg1, err := self.doTestPutFile()
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	msg1, err = self.doTestCheckSha1(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}
	return
}
