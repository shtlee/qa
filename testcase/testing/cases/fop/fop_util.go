package fop

import (
	"encoding/base64"
	"errors"
	"fmt"
	"qbox.me/sstore"
	"strconv"
	"strings"
)

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

func DecodeFh(efh string) *sstore.FhandleInfo {
	return sstore.DecodeFhandle(efh, "", KeyFinder)
}

func ExtractEfh(url string) string {
	idx := strings.LastIndex(url, "/")
	return string([]byte(url)[idx+1:])
}

func CookUrl(url, fopd string) string {
	efh := ExtractEfh(url)
	fhInfo := DecodeFh(efh)
	fsize := strconv.FormatInt(fhInfo.Fsize, 10)
	base64Fh := base64.URLEncoding.EncodeToString(fhInfo.Fhandle)
	reqUrl := fopd + "/op?fh=" + base64Fh + "&fsize=" + fsize + "&cmd="
	return reqUrl
}

func CheckImgInfo(server ImageInfo, local ImageInfo) error {
	switch {
	case server.Height != local.Height:
		return errors.New(fmt.Sprintf("Unmatched height! Expect : %d, But get %d \n",
			local.Height, server.Height))
	case server.Width != local.Width:
		return errors.New(fmt.Sprintf("Unmatched width! Expect : %d, But get %d \n",
			local.Width, server.Width))
	case server.ColorModel != local.ColorModel:
		return errors.New(fmt.Sprintf("Unmatched colorModel! Expect : %s, But get %s\n",
			local.ColorModel, server.ColorModel))
	case server.Format != local.Format:
		return errors.New(fmt.Sprintln("Unmatched format! Expect : %s, But get %s\n",
			local.Format, server.Format))
	default:
		return nil
	}
	return nil
}
