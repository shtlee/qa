package pub

import (
	"io"
	"os"
	"time"
	"fmt"
	"regexp"
	"strconv"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"math/rand"
	"path/filepath"
	"qbox.us/errors"
	"qbox.us/cc/config"
	da "qbox.me/auth/digest"
	"qbox.me/api"
	"qbox.me/api/rs"
	"qbox.me/api/pub"
	"qbox.me/api/util"
)

type Pub struct {
	Name               string  `json:"name"`
	Bucket             string  `json:"bucket"`
	Key                string  `json:"key"`
	DataFile           string `json:"data_file"`
	DataSha1 string `json:"data_sha1"`
	dataType string
	Domain string `json:"domain"`
	DomainIp string `json:"domain_ip"`
	isNormalDomain bool
	NormalDomainRegexp string `json:"normal_domain_regexp"`

	rsCli *rs.Service
	pubCli *pub.Service
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
	p.rsCli, err = rs.New(p.Env.Hosts, p.Env.Ips, dt)
	if err != nil {
		err = errors.Info(err, "Pub init failed")
		return
	}
	p.pubCli, err = pub.New(p.Env.Hosts["pu"], p.Env.Ips["pu"], dt)
	if err != nil {
		err = errors.Info(err, "Pub init failed")
		return
	}
	p.DataFile = filepath.Join(path, p.DataFile)
	domainRegexp, err := regexp.Compile(p.NormalDomainRegexp)
	if err != nil {
		err = errors.Info(err, "Pub init failed")
		return
	}
	p.isNormalDomain = domainRegexp.Match([]byte(p.Domain))

	return
}

func (p *Pub) doTestUpload() (msg string, err error) {

	p.dataType = "application/qbox-mon"
	entryName := p.Bucket + ":" + p.Key
	f, err := os.Open(p.DataFile)
	if err != nil {
		err = errors.Info(err, "Upload failed: ", p.DataFile)
		return
	}
	defer f.Close()
	fi, _ := f.Stat()
	begin := time.Now()
	_, _, err = p.rsCli.Put(entryName, p.dataType, f, fi.Size())
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("Pb    "+p.Env.Id+"_"+p.Name+"_doTestUpload", begin, end, duration)
	if err != nil {
		err = errors.Info(err, "upload failed:", entryName)
	}
	return
}

func (p *Pub) doTestPublish() (msg string, err error) {

	if p.isNormalDomain {
		p.Domain = p.Domain + "/" + strconv.FormatInt(rand.Int63(), 10)
	} else {
		p.Domain = strconv.FormatInt(rand.Int63(), 10) + "." + p.Domain
	}
	begin := time.Now()
	if _, err = p.rsCli.Publish(p.Domain, p.Bucket); err != nil {
		err = errors.Info(err, "Publish failed: ", p.Bucket, p.Domain)
		return
	}
	end := time.Now() 
	duration := end.Sub(begin)
	msg = util.GenLogEx("Pb    "+p.Env.Id+"_"+p.Name+"_doTestPublish", begin, end, duration)
	return
}

func (p *Pub) doTestDownload() (msg string, err error) {

	var (
		url string
	)
	if p.isNormalDomain {
		url = "http://" + p.Domain + "/" + p.Key
	} else {
		url = "http://" + p.DomainIp + "/" + p.Key
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = errors.Info(err, "Download failed:", url)
		return
	}
	if !p.isNormalDomain {
		req.Host = p.Domain
	}
	begin := time.Now()
	resp, err := http.DefaultClient.Do(req)
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("Fp    "+p.Env.Id+"_"+p.Name+"_doTestDownload", begin, end, duration)
	if err != nil {
		err = errors.Info(err, "Download failed:", url)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		err = errors.New("download status code is not 20x!")
		err = errors.Info(err, url, resp.StatusCode)
		return
	}
	if ct, ok := resp.Header["Content-Type"]; !ok || len(ct) != 1 || ct[0] != p.dataType {
		err = errors.New("Invalid content type")
		return
	}

	h := sha1.New()
	if _, err = io.Copy(h, resp.Body); err != nil {
		err = errors.Info(err, "check sha1 failed")
		return
	}
	if p.DataSha1 != hex.EncodeToString(h.Sum(nil)) {
		err = errors.New("Invalid data sha1")
		err = errors.Info(err, p.DataSha1, hex.EncodeToString(h.Sum(nil)))
		return
	}
	
	return
}


func (p *Pub) doTestUnpublish() (msg string, err error) {
	begin := time.Now()
	if _, err = p.rsCli.Unpublish(p.Domain); err != nil {
		err = errors.Info(err, "unpublish domain failed", p.Domain)
		return
	}
	end := time.Now()
	duration := end.Sub(begin)
	msg = util.GenLogEx("Pb    "+p.Env.Id+"_"+p.Name+"_doTestUnpublish", begin, end, duration)
	return
}


func (p *Pub) Test() (msg string, err error) {

	log1, err := p.doTestUpload()
	
	if err != nil {
		msg += fmt.Sprintln(log1, err)
		return
	} else {
		msg += fmt.Sprintln(log1, " ok")
	}

	log1, err = p.doTestPublish()
	if err != nil {
		msg += fmt.Sprintln(log1, err)
		return
	} else {
		msg += fmt.Sprintln(log1, " ok")
	}

	log1, err = p.doTestDownload()
	if err != nil {
		msg += fmt.Sprintln(log1, err)
		return
	} else {
		msg += fmt.Sprintln(log1, " ok")
	}

	log1, err = p.doTestUnpublish()
	if err != nil {
		msg += fmt.Sprintln(log1, err)
		return
	} else {
		msg += fmt.Sprintln(log1, " ok")
	}
	return
}