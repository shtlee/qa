package filepath

import (
	"io/ioutil"
	"os"
)

type Visitor interface {
	VisitDir(path string, fi os.FileInfo) bool
	VisitFile(path string, fi os.FileInfo)
}

func Walk(path string, v Visitor, error func(err error)) {

	fi, err := os.Lstat(path)
	if err != nil {
		if error != nil {
			error(err)
		}
		return
	}
	walk(path, fi, v, error)
}

func walk(path string, fi os.FileInfo, v Visitor, error func(err error)) {

	if !fi.IsDir() {
		v.VisitFile(path, fi)
		return
	}

	if !v.VisitDir(path, fi) {
		return // skip
	}

	fis, err := ioutil.ReadDir(path)
	if err != nil {
		if error != nil {
			error(err)
		}
		return
	}

	for _, e := range fis {
		walk(path+"/"+e.Name(), e, v, error)
	}
}
