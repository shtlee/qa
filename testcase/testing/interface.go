package main

import (
	"./cases/example"
	"./cases/fop"
	"./cases/up"
	"./cases/pub"
)


var (
	Cases = map[string]func () Interface{
		"example": func() Interface { return &example.Example{} },
		"resumable_put": func() Interface { return &up.UpPut{} },
		"up_put": func() Interface { return &up.UptokenPut{} },
		"up_put2": func() Interface { return &up.UpRPut{} },
		"fop_img_info":  func() Interface { return &fop.FopImgInfo{} },
		"fop_img_view":  func() Interface { return &fop.FopImgOp{} },
		"fop_img_mogr":  func() Interface { return &fop.FopImgOp{} },
		//"fop_img_exif":   &fop.FopImgExif{},
		//"old_mon":   &Old{},
		//"rs":          &Rs{},
		//"rs_upload":   &RsUpload{},
		"publish": func() Interface { return &pub.Pub{} },
		//"shell":       &Shell{},
		//"up":          &Up{},
	}
)

type Interface interface {
	Init(conf, env, path string) error
	Test() (msg string, err error)
}
