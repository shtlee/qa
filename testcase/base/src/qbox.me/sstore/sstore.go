package sstore

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"qbox.me/objid"
	"qbox.me/cc/time"
	"qbox.me/log"
)

// --------------------------------------------------------------------

type KeyFinder interface {
	Find(KeyHint uint32) []byte
}

// --------------------------------------------------------------------

type FhandleInfo struct {
	Fhandle  []byte // 321
	MimeType string // 32
	AttName  string // 32

	Fsize    int64 // 8
	Deadline int64 // 8

	KeyHint uint32 // 4
	OidLow  uint32 // 4
	OidHigh uint64 // 8

	Ver         uint16 // 2
	Compression uint8  // 1
	ChMask      uint8  // 1
	Uid         uint32 // 4

	Utype  uint32 // 4
	Public uint8  // 1
}

/*
	Token []byte		// 20

	Ver int8			// 1
	Reserved int8		// 1
	Compression int8	// 1
	ChMask int8			// 1

	ChunkCnt int8		// 1
	MimeTypeLen int8	// 1	
	AttNameLen int8		// 1
	Public int8			// 1

	KeyHint int32		// 4

	Fsize int64  		// 8
	Deadline int64		// 8

	Uid int32			// 4
	OidLow	int32		// 4
	OidHigh	int64		// 8

	Fhandle []byte
	MimeType string
	AttName string

	Utype int32			// 4
*/

const FhandleVer_10 = 0x10 // 1.0
const FhandleVer = 0x11    // 1.1

func EncodeFhandle_10(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 64 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName)
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	b[20] = FhandleVer_10
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)
	b[25] = byte(len(fi.MimeType))
	b[26] = byte(len(fi.AttName))
	b[27] = 0x27

	binary.LittleEndian.PutUint32(b[28:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[32:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[40:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[48:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[52:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[56:], fi.OidHigh^uint64(dwMask2))

	off := 64 + copy(b[64:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	for i := 64; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func DecodeFhandle_10(efh string, oid string, kf KeyFinder) *FhandleInfo {

	b, err := base64.URLEncoding.DecodeString(efh)
	if err != nil || len(b) < 64 {
		log.Println("DecodeString failed:", err)
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[40:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	fi.OidHigh = binary.LittleEndian.Uint64(b[56:]) ^ uint64(dwMask2)
	fi.OidLow = binary.LittleEndian.Uint32(b[52:]) ^ dwMask

	if fi.OidHigh != 0 && fi.OidLow != 0 {
		oidLow, oidHigh, err := objid.Decode(oid)
		if err != nil {
			log.Println("DecodeFhandle: objid.Decode failed")
			return nil
		}
		if fi.OidHigh != oidHigh || fi.OidLow != oidLow {
			log.Println("DecodeFhandle: SessionID not matched")
			return nil
		}
	}

	fi.KeyHint = binary.LittleEndian.Uint32(b[28:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer_10 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[32:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[48:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]

	for i := 64; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	off := len(b) - int(b[25]) - int(b[26])
	//	off := 64 + int(b[24])*20 + 1
	fi.Fhandle = b[64:off]

	off2 := off + int(b[25])
	fi.MimeType = string(b[off:off2])

	off = off2
	off2 = off + int(b[26])
	fi.AttName = string(b[off:off2])

	return fi
}

func EncodeFhandle(fi *FhandleInfo, key []byte) (efh string) {

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

	if fi.ChMask == 0 {
		fi.ChMask = uint8(fi.KeyHint ^ dwMask)
	}

	totalLen := 64 + len(fi.Fhandle) + len(fi.MimeType) + len(fi.AttName)
	if fi.Utype != 0 {
		totalLen += 4
	}
	chunkCnt := len(fi.Fhandle) / 20
	b := make([]byte, totalLen)

	if fi.Utype != 0 {
		b[20] = FhandleVer
	} else {
		b[20] = FhandleVer_10
	}
	b[21] = 0x34
	b[22] = fi.Compression
	b[23] = fi.ChMask

	b[24] = byte(chunkCnt)
	b[25] = byte(len(fi.MimeType))
	b[26] = byte(len(fi.AttName))
	b[27] = 0x27 ^ fi.Public

	binary.LittleEndian.PutUint32(b[28:], fi.KeyHint^dwMask)
	binary.LittleEndian.PutUint64(b[32:], uint64(fi.Fsize^dwMask2))
	binary.LittleEndian.PutUint64(b[40:], uint64(fi.Deadline))
	binary.LittleEndian.PutUint32(b[48:], fi.Uid^dwMask)
	binary.LittleEndian.PutUint32(b[52:], fi.OidLow^dwMask)
	binary.LittleEndian.PutUint64(b[56:], fi.OidHigh^uint64(dwMask2))

	off := 64 + copy(b[64:], fi.Fhandle)
	off += copy(b[off:], fi.MimeType)
	off += copy(b[off:], fi.AttName)

	if fi.Utype != 0 {
		binary.LittleEndian.PutUint32(b[off:], fi.Utype^dwMask)
		off += 4
	}

	for i := 64; i < off; i++ {
		b[i] ^= fi.ChMask
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	copy(b[:20], h.Sum(nil))

	log.Debug("Fhandle:\n" + hex.Dump(b))

	return base64.URLEncoding.EncodeToString(b)
}

func DecodeFhandle(efh string, oid string, kf KeyFinder) *FhandleInfo {

	b, err := base64.URLEncoding.DecodeString(efh)
	if err != nil || len(b) < 64 {
		log.Println("DecodeString failed:", err)
		return nil
	}

	fi := &FhandleInfo{}
	fi.Deadline = int64(binary.LittleEndian.Uint64(b[40:]))
	if fi.Deadline < time.Nanoseconds() {
		log.Println("Deadline expired")
		return nil
	}

	dwMask := uint32(fi.Deadline) ^ uint32(fi.Deadline>>32)
	dwMask2 := int64(dwMask) | (int64(dwMask) << 32)

/*	fi.OidHigh = binary.LittleEndian.Uint64(b[56:]) ^ uint64(dwMask2)
	fi.OidLow = binary.LittleEndian.Uint32(b[52:]) ^ dwMask

	if fi.OidHigh != 0 && fi.OidLow != 0 {
		oidLow, oidHigh, err := objid.Decode(oid)
		if err != nil {
			log.Println("DecodeFhandle: objid.Decode failed")
			return nil
		}
		if fi.OidHigh != oidHigh || fi.OidLow != oidLow {
			log.Println("DecodeFhandle: SessionID not matched")
			return nil
		}
	}
*/
	fi.KeyHint = binary.LittleEndian.Uint32(b[28:]) ^ dwMask
	key := kf.Find(fi.KeyHint)
	if key == nil {
		log.Println("KeyFinder: key not found")
		return nil
	}

	h := hmac.New(sha1.New, key)
	h.Write(b[20:])
	if !bytes.Equal(h.Sum(nil), b[:20]) {
		log.Println("Digest verify failed")
		return nil
	}

	fi.Ver = uint16(b[20])
	if fi.Ver != FhandleVer && fi.Ver != FhandleVer_10 {
		log.Println("FhandleVer not match")
		return nil
	}

	fi.Fsize = int64(binary.LittleEndian.Uint64(b[32:])) ^ dwMask2
	fi.Uid = binary.LittleEndian.Uint32(b[48:]) ^ dwMask

	fi.Compression = b[22]
	fi.ChMask = b[23]
	fi.Public = b[27] ^ 0x27

	for i := 64; i < len(b); i++ {
		b[i] ^= fi.ChMask
	}

	//	off := 64 + int(b[24])*20 + 1
	off := len(b) - int(b[25]) - int(b[26])
	if fi.Ver == FhandleVer {
		off -= 4
	}

	fi.Fhandle = b[64:off]

	off2 := off + int(b[25])
	fi.MimeType = string(b[off:off2])

	off = off2
	off2 = off + int(b[26])
	fi.AttName = string(b[off:off2])

	if fi.Ver == FhandleVer {
		fi.Utype = binary.LittleEndian.Uint32(b[off2:]) ^ dwMask
	}

	return fi
}

// --------------------------------------------------------------------
