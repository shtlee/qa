package  util

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"bytes"
	"io"
	"net/http"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "code.google.com/p/go.image/tiff"
	"qbox.me/sstore"
	"strconv"
	"encoding/base64"
)


// UP ==>> env.id + casename + func_name + begin + end + duration
func GenLog(msg string, begin, end time.Time, duration time.Duration) string {
	sBegin := begin.String()
	msIdx := 23
	sBegin = (string)([]byte(sBegin)[10:msIdx])

	sEnd := end.String()
	sEnd = (string)([]byte(sEnd)[10:msIdx])

	sDuration := duration.String()
	dotIdx := strings.LastIndex(sDuration, ".")
	sDuration = (string)([]byte(sDuration)[:dotIdx+2]) + "ms"
	return fmt.Sprintf("%-45s %-15s %-15s %8s", msg, sBegin, sEnd, sDuration)
}

func DoHttpGet(url string) (b *bytes.Buffer, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}

	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}

	defer resp.Body.Close()
	b = new(bytes.Buffer)
	io.Copy(b, resp.Body)
	return
}


func approachTo(a1 uint32, a2 uint32) bool {
	var max, d int
	if a1 < a2 {
		max = int(a2)
		d = int(a2 - a1)
	} else {
		max = int(a1)
		d = int(a1 - a2)
	}
	if d <= max/10 {
		return true
	}
	return false
}

func CheckImg(src, tgt io.Reader) (int, error) {
	image1, format1, err1 := image.Decode(src)
	image2, format2, err2 := image.Decode(tgt)
	if err1 != nil || err2 != nil {
		return 1, errors.New(fmt.Sprintf("Server error : %v , Local error : %v \n", err1, err2))
	}
	if format2 != format1 {
		return 2, errors.New("Unmatched format!")
	}
	if image1.Bounds() != image2.Bounds() {
		return 3, errors.New("Unmatched bounds")
	}

	total, miss := 0, 0
	for i := image1.Bounds().Min.X; i <= image1.Bounds().Max.X; i++ {
		for j := image1.Bounds().Min.Y; j <= image1.Bounds().Max.Y; j++ {
			total++
			r1, g1, b1, a1 := image1.At(i, j).RGBA()
			r2, g2, b2, a2 := image2.At(i, j).RGBA()
			if !approachTo(r1, r2) || !approachTo(g1, g2) || !approachTo(b1, b2) || !approachTo(a1, a2) {
				return 4, errors.New("Differs two much!")
			}
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				miss++
			}
		}
	}
	if miss > total/20 {
		return 4, errors.New("Missed too much!")
	}
	return 0, nil
}


const KeyHintESTest = 139
const KeyHintMockFS = 113
const KeyHintFSTest = 103
const KeyHintIOTest = 104
const KeyHintRSTest = 105
const KeyHintPubTest = 106

var KeyMockFS = []byte("qbox.mockfs")
var KeyFSTest = []byte("qbox.fs.test")
var KeyIOTest = []byte("qbox.io.test")
var KeyESTest = []byte("qbox.es.test")
var KeyRSTest = []byte("qbox.rs.test")
var KeyPubTest = []byte("qbox.pub.test")
var KeyFinder = sstore.SimpleKeyFinder(map[uint32][]byte{
	KeyHintMockFS:  KeyMockFS,
	KeyHintFSTest:  KeyFSTest,
	KeyHintIOTest:  KeyIOTest,
	KeyHintESTest:  KeyESTest,
	KeyHintRSTest:  KeyRSTest,
	KeyHintPubTest: KeyPubTest,
})

func decodeFh(efh string) *sstore.FhandleInfo {
	return sstore.DecodeFhandle(efh, "", KeyFinder)
}
func CookUrl(url, fopd string) string {
	efh := extractEfh(url)
	fhInfo := decodeFh(efh)
	fsize := strconv.FormatInt(fhInfo.Fsize, 10)
	base64Fh := base64.URLEncoding.EncodeToString(fhInfo.Fhandle)
	reqUrl := fopd + "/op?fh=" + base64Fh + "&fsize=" + fsize + "&cmd="
	return reqUrl
}



func extractEfh(url string) string {
	idx := strings.LastIndex(url, "/")
	return string([]byte(url)[idx+1:])
}



/*
func GetAuthWithKey(accHost string, user, password string, key string) (t http.RoundTripper, err error) {
	cfg := &oauth.Config{
		ClientId:     key,
		ClientSecret: "<ClientSecret>",
		Scope:        "<Scope>",
		AuthURL:      "<AuthURL>",
		TokenURL:     accHost + "/oauth2/token",
		RedirectURL:  "<RedirectURL>",
	}

	transport := &oauth.Transport{Config: cfg}
	_, _, err = transport.ExchangeByPassword(user, password)
	t = transport
	return
}

func GetAuth(accHost string, user, password string) (t http.RoundTripper, err error) {
	return GetAuthWithKey(accHost, user, password, "abcd0c7edcdf914228ed8aa7c6cee2f2bc6155e2")
}
*/