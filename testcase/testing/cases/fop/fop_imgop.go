package fop

import (
	"fmt"
	"time"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"qbox.us/cc/config"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/util"
	da "qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
)

type FopImgOp struct {
	Name       string `json:"name"`
	BucketName string `json:"bucket"`
	Key        string `json:"key"`

	ChunkSize int  `json:"chunk_size"`
	BlockBits uint `json:"block_bits"`

	SrcImg    string `json:"source_file"`
	TargetImg string `json:"target_file"`
	Op        string `json:"op"`

	Env  api.Env
}

func (self *FopImgOp) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return
	}
	self.SrcImg = filepath.Join(path, self.SrcImg)
	self.TargetImg = filepath.Join(path, self.TargetImg)
	return
}

// upload the file and get the download url 
func (self *FopImgOp) doTestGetImgUrl() (url string, err error) {
	entry := self.BucketName + ":" + self.Key

	dt := da.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
	rsservice, err := rs.New(self.Env.Hosts, self.Env.Ips, dt)
	if err != nil {
		return
	}
	authPolicy := &uptoken.AuthPolicy{
		Scope:    entry,
		Deadline: 3600,
	}
	authPolicy.Deadline += uint32(time.Now().Unix())
	token := uptoken.MakeAuthTokenString(self.Env.AccessKey, self.Env.SecretKey, authPolicy)
	_, code, err := rsservice.Upload(entry, self.SrcImg, "", "", "", token)

	if err != nil || code != 200 {
		return
	}

	getRet, code, err := rsservice.Get(entry, "", "", 3600)
	if err != nil || code != 200 {
		return
	}

	url = getRet.URL
	return
}

func (self *FopImgOp) doTestImgOp(downloadUrl string) (msg string, err error) {
	begin := time.Now()
	url := downloadUrl + self.Op
	netBuf, err := util.DoHttpGet(url)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("Fp    "+self.Env.Id+"_"+self.Name+"_doTestImgOp", begin, end, duration)
	if err != nil {
		return
	}
	targetFile, err := os.Open(self.TargetImg)
	if err != nil {
		return
	}
	_, err = util.CheckImg(netBuf, targetFile)
	if err != nil {
		return
	}

	return
}

func (self *FopImgOp) Test() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgUrl()
	if err != nil {
		return
	}

	url = util.CookUrl(url, self.Env.Fopd)
	msg1, err = self.doTestImgOp(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	return
}
