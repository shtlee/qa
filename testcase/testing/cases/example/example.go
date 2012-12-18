package example 

import (
	"errors"
	"qbox.us/cc/config"
)

type ExampleConf struct {
	Msg string `json:"msg"`
	Err bool   `json:"err"`
}

type Example struct {
	conf *ExampleConf
}

func (p *Example) Init(conf, env, path string) (err error) {

	if err = config.LoadEx(&p.conf, conf); err != nil {
		return
	}
	return nil
}

func (p *Example) Test() (msg string, err error) {
	if p.conf.Err {
		return p.conf.Msg, errors.New("example err")
	}
	return p.conf.Msg, nil
}
