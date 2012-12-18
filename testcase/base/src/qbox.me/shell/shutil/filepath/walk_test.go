package filepath

import (
	"fmt"
	"os"
	"testing"
)

type visitor struct {
}

func (v visitor) VisitDir(path string, f os.FileInfo) bool {
	fmt.Println("dir:", path)
	return true
}

func (v visitor) VisitFile(path string, f os.FileInfo) {
	fmt.Println("file:", path)
}

func TestWalk(t *testing.T) {
	Walk("c:/./QBox\\f", visitor{}, nil)
}
