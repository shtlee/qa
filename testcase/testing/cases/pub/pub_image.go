package pub

import (
	"qbox.me/api"	
	"qbox.me/api/pub"
)

type PubImage struct {

	Name string `json:"name"`
	Bucket string `json:"bucket"`
	Key string `json:"key"`
	DataFile string `json:"data_file"`
	DataSha1 string `json:"data_sha1"`
	Domain string `json:"image_domain"`
	ImageURL string `json:"source_image_url"`
	
	Env api.Env
}

func (p *Pub) Init(conf, env, path string) (err error) {

}

func (p *Pub) Test() (msg string, err error) {

	
}