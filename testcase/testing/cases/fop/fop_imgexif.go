package fop

import (
	"time"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/util"
	"qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
	"qbox.us/cc/config"
)

type FopImgExif struct {

	Name       string `json:"name"`
	BucketName string `json:"bucket"`
	Key        string `json:"key"`

	ChunkSize int  `json:"chunk_size"`
	BlockBits uint `json:"block_bits"`

	UploadImg string `json:"img_data"`
	SrcExif string `json:"source_exif"`

	Env  api.Env
}

func (self *FopImgExif) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return
	}
	self.UploadImg = filepath.Join(path, self.UploadImg)
	self.SrcExif = filepath.Join(path, self.SrcExif)
	return
}

// upload the file and get the download url 
func (self *FopImgExif) doTestGetImgUrl() (url string, err error) {
	entry := self.BucketName + ":" + self.Key

	dt := digest.NewTransport(self.Env.AccessKey, self.Env.SecretKey, nil)
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
	_, code, err := rsservice.Upload(entry, self.UploadImg, "", "", "", token)

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

func (self *FopImgExif) doTestImgExif(downloadUrl string) (msg string, err error) {
	begin := time.Now()
	url := downloadUrl + "exif"
	netBuf, err := util.DoHttpGet(url)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("Fp    "+self.Env.Id+"_"+self.Name+"_doTestImgOp", begin, end, duration)
	if err != nil {
		return
	}
	fmt.Println("-----------------------> netbuf : ", netBuf)
	return
}

func (self *FopImgExif) Test() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgUrl()
	if err != nil {
		return
	}

	url = util.CookUrl(url, self.Env.Fopd)
	fmt.Println("url ---------png------> ", url)
	msg1, err = self.doTestImgExif(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	return
}
