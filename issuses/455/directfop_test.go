package direct_fop

import (
	"os"
	"io/ioutil"
	"net/http"
	"testing"
	"qbox.us/rpc"
	"qbox.me/api/rs"
	"qbox.us/digest_auth"
)

var (
	rsCli *rs.Service
	cli *http.Client
	Accesskey = ""
	Secretkey = ""
	Bucket = ""
	Key = ""
	URL = ""
	TestFile = ""
	Hosts = map[string]string{
		"rs": "192.168.1.201",
	}
	Ips = map[string]string{
		"rs": "http://192.168.1.201:9400",
	}
)

func Init() (err error) {

	dt := digest_auth.NewTransport(Accesskey, Secretkey, nil)
	rsCli, err = rs.New(Hosts, Ips, dt)
	if err != nil {
		return
	}
	cli = &http.Client{
		Transport: dt,
	}
	return
}

func doTestUpload(t *testing.T) {

	f, err := os.Open(TestFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	entryURI := Bucket + ":" + Key
	_, _, err = rsCli.Put(entryURI, "", f, fi.Size())
	if err != nil {
		t.Fatal(err)
	}
	ret, _, err := rsCli.Get(entryURI, "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	URL = ret.URL
	t.Log(URL)
}

func doTestSaveAs(t *testing.T) {

	encodeEURI := rpc.EncodeURI(Bucket + Key)
	url := URL + "?save-as/" + encodeEURI + "/imageMogr/format/bmp"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestSaveAs2(t *testing.T) {

	encodeEURI := rpc.EncodeURI(Bucket + Key)
	url := URL + "?iamgeMogr/format/bmp/save-as/" + encodeEURI
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestImageMogr(t *testing.T) {

	url := URL + "?imageMogr/format/bmp"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestExif(t *testing.T) {

	url := URL + "?exif"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func doTestImageInfo(t *testing.T) {

	url := URL + "?imageInfo"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func doTestUrlInfo(t *testing.T) {

	url := URL + "?urlInfo"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func doTestImageView(t *testing.T) {

	url := URL + "?imageView/0/w/101"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestAvthumb(t *testing.T) {

	url := URL + "?avthumb/xxx"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestImagePreview(t *testing.T) {

	url := URL + "?imagePreview/24"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestImagePreviewEx(t *testing.T) {

	url := URL + "?imagePreviewEx/24"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}

func doTestStat(t *testing.T) {

	url := URL + "?stat"
	resp, err := cli.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Log(resp)
}


func TestDo(t *testing.T) {

	if err := Init(); err != nil {
		t.Fatal(err)
	}
	doTestUpload(t)
	doTestSaveAs(t)
	doTestImageMogr(t)
	doTestExif(t)
	doTestImageInfo(t)
	doTestUrlInfo(t)
	doTestImageView(t)

	doTestSaveAs2(t)
	doTestImagePreview(t)
	doTestImagePreviewEx(t)
	doTestStat(t)

	return
}