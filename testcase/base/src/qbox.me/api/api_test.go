package api

import (
	"testing"
	"io/ioutil"
	"math/rand"
	"qbox.me/auth/digest"
	"qbox.me/api/up2"
	"qbox.me/api/rs"
)


var (
	hosts = map[string]string{
		"io": "iovip.qbox.me",
		"rs": "rs.qbox.me",
		"up": "up.qbox.me",
		"eu": "eu.qbox.me",
		"pu": "pu.qbox.me",
		"uc": "uc.qbox.me",
	}
	ips = map[string]string{
		"io": "http://iovip.qbox.me",
		"rs": "http://rs.qbox.me",
		"up": "http://up.qbox.me",
		"eu": "http://eu.qbox.me",
		"pu": "http://pu.qbox.me:10200",
		"uc": "http://uc.qbox.me",
	}
	BlockBits = uint(22)
	RPutChunkSize = 262144
	RPutRetryTimes = 2
	AccessKey = "4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_kO"
	SecretKey = "SrRuUVfDX6drVRvpyN8mv8Vcm9XnMZzlbDfvVfmE"

	testbucket = "test_bucket"
	testkey = "gopher.jpg"
)

var (
	Rscli *rs.Service
	Up2cli *up2.Service
)

func Init() (err error) {

	dt := digest.NewTransport(AccessKey, SecretKey, nil)
	Up2cli, err = up2.New(hosts["up"], ips["up"], BlockBits, RPutChunkSize, RPutRetryTimes, dt)
	if err != nil {
		return
	}
	Rscli, err = rs.New(hosts, ips, dt)
	if err != nil {
		return
	}
	return
}


// Advance resumable put
func doTestRPutTask(t *testing.T) {

	tfsize := int64(10*1024*1024)
	tf, err := ioutil.TempFile("","RPutTest")
	if err != nil {
		t.Fatal(err)
	}
	defer tf.Close()
	if err = tf.Truncate(tfsize); err != nil {
		t.Fatal(err)
	}

	entryURI := testbucket + ":" + testkey + "1"
	blockcnt := int((tfsize + (1 << Up2cli.BlockBits) - 1) >> Up2cli.BlockBits)
	progs := make([]up2.BlockputProgress, blockcnt)

	chunkNotify := func(idx int, p *up2.BlockputProgress) {
		if rand.Intn(blockcnt)/3 == 0  {
			p1 := *p
			progs[idx] = p1
		}
	}

	blockNotify := func(idx int, p *up2.BlockputProgress) {
	}
	t1 := Up2cli.NewRPtask(entryURI, "", "", "", "", tf, tfsize, nil)
	t1.ChunkNotify = chunkNotify
	t1.BlockNotify = blockNotify
	for i := 0; i < blockcnt; i++ {
		t1.PutBlock(i)
	}
	t.Log(progs)
	t1.Progress = progs
	code, err := t1.Run(10, 10, nil, nil)
	if err != nil || code/100 != 2 {
		t.Fatal("Resumable put failed", err, code)
	}
	
	entry, code, err := Rscli.Get(entryURI, "", "", 0)
	if err != nil || code/100 != 2 {
		t.Fatal("Get failed", err, code)
	}
	t.Log(entry)
	code, err = Rscli.Delete(entryURI)
	if err != nil || code/100 != 2 {
		t.Fatal("Delete failed", err, code)
	}
}

func TestDo(t *testing.T) {

	if err := Init(); err != nil {
		t.Fatal(err)
	}
	doTestRPutTask(t)
}