package pub

import (
	"path/filepath"
	"qbox.us/cc/config"
	da "qbox.me/auth/digest"
	"qbox.me/api"
	"qbox.me/api/pub"
)

type Pub struct {
	Name               string
	Bucket             string
	Key                string
	DataFile           string
	NormalDomainRegexp string

	Conn *pub.Service
	Env api.Env
}

func (p *Pub) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(p, conf); err != nil {
		return
	}
	if err = config.LoadEx(&p.Env, env); err != nil {
		return
	}
	dt := da.NewTransport(p.Env.AccessKey, p.Env.SecretKey, nil)
	p.Conn, err = pub.New(p.Env.Hosts["pub"], p.Env.Ips["pub"], dt)
	p.DataFile = filepath.Join(path, p.DataFile)
	return
}

func (p *Pub) doTestPublish() (msg string, err error) {

	return
}
