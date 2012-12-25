package fop

import (
	"encoding/json"
	"errors"
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
	"time"
)

type FopImgExif struct {
	Name       string `json:"name"`
	BucketName string `json:"bucket"`
	Key        string `json:"key"`

	ChunkSize int  `json:"chunk_size"`
	BlockBits uint `json:"block_bits"`

	UploadImg string `json:"img_data"`
	SrcExif   ImgExif
	Env       api.Env
}

type ValueTypePair struct {
	Value string `json:"val"`
	Type  int    `json:"type"`
}

type ImgExif struct {
	ColorSpace              ValueTypePair `json:"ColorSpace"`
	ComponentsConfiguration ValueTypePair `json:"ComponentsConfiguration"`
	CompressedBitsPerPixel  ValueTypePair `json:"CompressedBitsPerPixel"`
	Compression             ValueTypePair `json:"Compression"`
	Contrast                ValueTypePair `json:"Contrast"`
	CustomRendered          ValueTypePair `json:"CustomRendered"`
	DateTime                ValueTypePair `json:"DateTime"`
	DateTimeDigitized       ValueTypePair `json:"DateTimeDigitized"`
	DateTimeOriginal        ValueTypePair `json:"DateTimeOriginal"`
	ExifVersion             ValueTypePair `json:"ExifVersion"`
	ExposureBiasValue       ValueTypePair `json:"ExposureBiasValue"`
	ExposureMode            ValueTypePair `json:"ExposureMode"`
	ExposureProgram         ValueTypePair `json:"ExposureProgram"`
	ExposureTime            ValueTypePair `json:"ExposureTime"`
	FNumber                 ValueTypePair `json:"FNumber"`
	FileSource              ValueTypePair `json:"FileSource"`
	Flash                   ValueTypePair `json:"Flash"`
	FlashPixVersion         ValueTypePair `json:"FlashPixVersion"`
	FocalLength             ValueTypePair `json:"FocalLength"`
	ISOSpeedRatings         ValueTypePair `json:"ISOSpeedRatings"`
	ImageDescription        ValueTypePair `json:"ImageDescription"`
	LightSource             ValueTypePair `json:"LightSource"`
	Make                    ValueTypePair `json:"Make"`
	MaxApertureValue        ValueTypePair `json:"MaxApertureValue"`
	MeteringMode            ValueTypePair `json:"MeteringMode"`
	Model                   ValueTypePair `json:"Model"`
	Orientation             ValueTypePair `json:"Orientation"`
	PixelXDimension         ValueTypePair `json:"PixelXDimension"`
	PixelYDimension         ValueTypePair `json:"PixelYDimension"`
	PrintImageMatching      ValueTypePair `json:"PrintImageMatching"`
	ResolutionUnit          ValueTypePair `json:"ResolutionUnit"`
	Saturation              ValueTypePair `json:"Saturation"`
	SceneCaptureType        ValueTypePair `json:"SceneCaptureType"`
	SceneType               ValueTypePair `json:"SceneType"`
	Sharpness               ValueTypePair `json:"Sharpness"`
	Software                ValueTypePair `json:"Software"`
	WhiteBalance            ValueTypePair `json:"WhiteBalance"`
	XResolution             ValueTypePair `json:"XResolution"`
	YCbCrPositioning        ValueTypePair `json:"YCbCrPositioning"`
	YResolution             ValueTypePair `json:"YResolution"`
}

func (self *FopImgExif) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(self, conf); err != nil {
		return
	}
	if err = config.LoadEx(&self.Env, env); err != nil {
		return
	}
	if err = config.LoadEx(&self.SrcExif, conf); err != nil {
		return
	}
	self.UploadImg = filepath.Join(path, self.UploadImg)
	//self.SrcExif = filepath.Join(path, self.SrcExif)
	fmt.Println("*********************************************> ", self.SrcExif)
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
	var TargetExif ImgExif
	json.Unmarshal(netBuf.Bytes(), &TargetExif)
	if self.SrcExif != TargetExif {
		err = errors.New("Umatched Exif!")
	}
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("Fp    "+self.Env.Id+"_"+self.Name+"_doTestImgOp", begin, end, duration)
	if err != nil {
		return
	}
	return
}

func (self *FopImgExif) Test() (msg string, err error) {
	msg1 := ""
	url, err := self.doTestGetImgUrl()
	if err != nil {
		return
	}

	url = util.CookUrl(url, self.Env.Fopd)
	msg1, err = self.doTestImgExif(url)
	if err == nil {
		msg += fmt.Sprintln(msg1, " ok")
	} else {
		msg += fmt.Sprintln(msg1, err)
	}

	return
}
