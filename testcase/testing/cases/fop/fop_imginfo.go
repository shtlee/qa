package fop

import (
	"fmt"
	"time"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"encoding/json"
	"path/filepath"
	"qbox.us/log"
	"qbox.us/cc/config"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/util"
	da "qbox.me/auth/digest"
	"qbox.me/auth/uptoken"
)

type FopImgInfo struct {
	// Ops []string `json:"ops"`
	Name string `json:"name"`

	ImageFile     string    `json:"image_file"`
	BucketName    string    `json:"bucket"`
	Key           string    `json:"key"`
	ChunkSize     int       `json:"chunk_size"`
	BlockBits     uint      `json:"block_bits"`
	SrcImg        string    `json:"source_file"`
	TargetImgInfo ImageInfo `json:"targe_imginfo"`
	Fopd          string    `json:"fopd"`
	FopdLogFile   string    `json:"log_fopd"`
	FopdLoger     *log.Logger

	Env api.Env
}

type ImageInfo struct {
	Format     string `json:"format"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	ColorModel string `json:"colorModel"`
}

func (self *FopImgInfo) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return
	}
	self.SrcImg = filepath.Join(path, self.SrcImg)
	return
}

// upload the file and get the download url 
func (self *FopImgInfo) doTestGetImgUrl() (url string, err error) {
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

func (self *FopImgInfo) doTestGetImgInfo(downloadUrl string) (msg string, err error) {
	begin := time.Now()
	url := downloadUrl + "imageInfo"
	netBuf, err := util.DoHttpGet(url)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLog("Fp    "+self.Env.Id+"_"+self.Name+"_doTestGetImgInfo", begin, end, duration)
	if err != nil {
		return
	}

	var serImgInfo ImageInfo
	json.Unmarshal(netBuf.Bytes(), &serImgInfo)
	locImgInfo := self.TargetImgInfo

	if err = CheckImgInfo(serImgInfo, locImgInfo); err != nil {
		return
	}
	return
}

func (self *FopImgInfo) Test() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgUrl()
	if err != nil {
		return
	}
	url = util.CookUrl(url, self.Env.Fopd)
	msg1, err = self.doTestGetImgInfo(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	return
}
