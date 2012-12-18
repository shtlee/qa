package main

import (
	"./cases/example"
	"./cases/fop"
	"./cases/up"
)

type Interface interface {
	Init(conf, env, path string) error
	Test() (msg string, err error)
}

var Cases = map[string]Interface{
	"example":       &example.Example{},
	"resumable_put": &up.UpResuPut{},
	"fop_img_info":  &fop.FopImgInfo{},
	"fop_img_view":  &fop.FopImgOp{},
	"fop_img_mogr":  &fop.FopImgOp{},
	//"fop_img_exif"        :    &fop.FopImgExif{},
	"up_put": &up.PutFile{},
	/*"old_mon":   &Old{},
	"rs":          &Rs{},
	"rs_upload":   &RsUpload{},
	"publish":     &Publish{},
	"shell":       &Shell{},
	"up":          &Up{},*/
}
