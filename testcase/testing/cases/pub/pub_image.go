package pub

import (
	"qbox.us/cc/config"
	"qbox.us/errors"
	"qbox.me/api"	
	"qbox.me/api/pub"
)

type PubImage struct {

	Name string `json:"name"`
	Bucket string `json:"bucket"`
	Domain string `json:"image_domain"`
	FileURL string `json:"source_file_url"`

	Pubcli *pub.Service
	Env api.Env
}

func (p *Pub) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(p, conf); err != nil {
		err = errors.Info(err, "pub_image load conf failed", conf)
		return
	}
	if err = config.LoadEx(&p.Env, env); err != nil {
		err = errors.Info(err, "pub_image load env failed", env)
		return
	}
	dt := da.NewTransport(p.Env.AccessKey, p.Env.SecretKey, nil)
	p.Pubcli, err = pub.New(p.Env.Hosts["pub"], p.Env.Ips["pub"], dt)
	if err != nil {
		err = errors.Info(err, "pub_image init failed")
		return
	}
	return
}

func (p *Pub) doTestImage() (msg string, err error) {

	code, err := p.Pubcli.Image(p.Bucket, )
}


func (p *Pub) Test() (msg string, err error) {

	
}