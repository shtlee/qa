package main

import (
	"flag"
	"fmt"
	"os"
	filepath1 "path/filepath"
	"qbox.me/shell/shutil/filepath"
	"qbox.us/cc"
	"qbox.us/cc/config"
	"qbox.us/log"
	"runtime"
)

type Config struct {
	MaxProcs int    `json:"max_procs"`
	DataPath string `json:"data"`
	Include  string `json:"cases"`
	Env      string `json:"env"`
}

type Visitor struct {
	*Config
	cases map[string]Interface
}

type CaseInfo struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Enable bool   `json:"enable"`
}

func (p *Visitor) VisitDir(path string, fi os.FileInfo) bool { return true }
func (p *Visitor) VisitFile(file string, fi os.FileInfo) {

	var (
		conf CaseInfo
	)
	log.Info("loading ", file, "...")
	if err := config.LoadEx(&conf, file); err != nil {
		log.Error("load err :", file, err)
		os.Exit(1)
	}
	if conf.Enable == true {
		if _, ok := p.cases[conf.Name]; ok || conf.Name == "" {
			log.Error("name err or duplicate: ", file, conf.Name, conf.Type)
			os.Exit(-1)
		}
		caseEntry, ok := Cases[conf.Type]
		if !ok {
			log.Error("no such type :", conf.Type, conf.Name)
			os.Exit(1)
		}
		err := caseEntry.Init(file, p.Env, p.DataPath)
		if err != nil {
			log.Error("init err :", conf.Name, conf.Type, err)
			os.Exit(1)
		}
		p.cases[conf.Name] = caseEntry
		log.Info("loaded", conf.Name, conf.Type)
	}
}

func main() {
	var (
		conf Config
	)
	confDir, _ := cc.GetConfigDir("qbox.me")
	var confName *string = flag.String("f", confDir+"/qboxtest.conf", "the config file")
	flag.Parse()
	log.Info("Use the config file of " + *confName)
	if err := config.LoadEx(&conf, *confName); err != nil {
		log.Error(err)
		return
	}

	runtime.GOMAXPROCS(conf.MaxProcs)

	cases := make(map[string]Interface)
	conf.Include = filepath1.Join(confDir, conf.Include)
	conf.DataPath = filepath1.Join(confDir, conf.DataPath)
	conf.Env = filepath1.Join(confDir, conf.Env)
	filepath.Walk(conf.Include, &Visitor{&conf, cases}, nil)

	check := func() bool {
		msg := fmt.Sprintf("begin check ...\n")
		errCount := 0
		count := 0
		for k, v := range cases {
			count++
			log.Info("begin check", k, "...")
			msg += fmt.Sprintf("[%v/%v]process %v <<<\n", count, len(cases), k)
			info, err := v.Test()
			log.Info("check done :\n", info, k, err)
			msg += info
			msg += "\n"
			if err != nil {
				msg += fmt.Sprintf("!!!!!!!!!!mon case [%v] err!!![%v]\n", k, err)
				errCount++
			} else {
				msg += fmt.Sprintf("[no err]%v done <<<\n", k)
			}
		}
		msg += fmt.Sprintf("all cases finish[%v/%v] <<<\n", len(cases)-errCount, len(cases))
		fmt.Println("----------------- result ------------------")
		fmt.Println(msg)
		fmt.Println("-------------------------------------------")
		return errCount == 0
	}

	b := check()
	if b {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
