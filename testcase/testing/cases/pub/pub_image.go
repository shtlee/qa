package pub

import (
	"qbox.us/cc/config"
	"qbox.us/errors"
	"qbox.me/api"	
	"qbox.me/api/pub"
	"qbox.me/httputil"
	da "qbox.me/auth/digest"
)

type PubImage struct {

	Name string `json:"name"`
	Bucket string `json:"bucket"`
	FromDomain string `json:"from_domain"`
	SrcHost string `json:"source_host"`
	SrcKey string `json:"source_key"`

	Pubcli *pub.Service
	Env api.Env
}

func (p *PubImage) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(p, conf); err != nil {
		err = errors.Info(err, "pub_image load conf failed", conf)
		return
	}
	if err = config.LoadEx(&p.Env, env); err != nil {
		err = errors.Info(err, "pub_image load env failed", env)
		return
	}
	dt := da.NewTransport(p.Env.AccessKey, p.Env.SecretKey, nil)
	p.Pubcli, err = pub.New(p.Env.Hosts["pu"], p.Env.Ips["pu"], dt)
	if err != nil {
		err = errors.Info(err, "pub_image init failed")
		return
	}
	return
}

func (p *PubImage) doTestImage() (msg string, err error) {

	from := []string{p.FromDomain}
	code, err := p.Pubcli.Image(p.Bucket, from, p.SrcHost, 0)
	if err != nil || code/100 != 2 {
		if err == nil {
			err = errors.New("doTestImage failed")
		} 
		err = errors.Info(err, code, p.SrcHost, p.FromDomain)
		return
	}
	url := "http://" + p.Env.Hosts["io"] + "/" + p.SrcKey
	_, err = httputil.DownloadEx(url, p.FromDomain)
	if err != nil {
		err = errors.Info(err, "doTestImage failed", url)
		return
	}
	return
}

func (p *PubImage) doTestUnimage() (msg string, err error) {

	code, err := p.Pubcli.Unimage(p.Bucket)
	if err != nil || code/100 != 2 {
		if err == nil {
			err = errors.New("doTestUnimage failed")
		}
		err = errors.Info(err, code, p.Bucket)
		return
	}
	return
}

func (p *PubImage) Test() (msg string, err error) {

	msg1, err := p.doTestImage()
	if err != nil {
		return
	}
	msg += msg1

	msg2, err := p.doTestUnimage()
	if err != nil {
		return
	}
	msg += msg2
	
	return
}