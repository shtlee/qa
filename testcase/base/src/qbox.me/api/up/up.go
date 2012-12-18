package up

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"qbox.me/httputil"
	"strconv"
	"sync"
)

const (
	InvalidCtx      = 701
	BLOCK_BITS      = 22
	PUT_CHUNK_SIZE  = 256
	PUT_RETRY_TIMES = 3
)

type PutRet struct {
	Ctx      string `json:"ctx"`
	Checksum string `json:"checksum"`
	Crc32    uint32 `json:"crc32"`
	Offset   uint32 `json:"offset"`
	Host     string `json:"host"`
}

type Service struct {
	host, ip       string
	BlockBits      uint
	PutChunkSize   int
	RPutRetryTimes int
	Tasks          chan func()
	Conn           *httputil.Client
}

func NewService(host, ip string, blockbits uint, chunksize,
	retryTimes int, t http.RoundTripper, taskQsize, threadSize int) (s Service, err error) {
	tasks := make(chan func(), taskQsize)
	for i := 0; i < threadSize; i++ {
		go worker(tasks)
	}
	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = Service{host, ip, blockbits, chunksize, retryTimes, tasks, &httputil.Client{client}}
	return
}

func (r Service) mkBlock(blockSize int, body io.Reader, bodyLength int) (ret PutRet, code int, err error) {
	code, err = r.Conn.CallWithEx(
		&ret, r.ip+"/mkblk/"+strconv.Itoa(blockSize), r.host, "application/octet-stream", body, (int64)(bodyLength))
	return
}

func (r Service) blockPut(ctx string, offset int, body io.Reader, bodyLength int) (ret PutRet, code int, err error) {
	code, err = r.Conn.CallWithEx(
		&ret, r.ip+"/bput/"+ctx+"/"+strconv.Itoa(offset), r.host, "application/octet-stream", body, (int64)(bodyLength))
	return
}

type BlockProgress struct {
	Ctx      string
	Offset   int
	RestSize int
	Err      error
}

func blockFirstPut(prog *BlockProgress) bool {
	return prog.Ctx == ""
}

func getBodyLength(chunkSize, blkSize int) int {
	var bodyLength int
	if chunkSize < blkSize {
		bodyLength = chunkSize
	} else {
		bodyLength = blkSize
	}
	return bodyLength
}

func (r Service) resumableBlockput(
	f io.ReaderAt, blockIdx int, blkSize, chunkSize, retryTimes int,
	prog *BlockProgress, notify func(blockIdx int, prog *BlockProgress)) (ret PutRet, code int, err error) {

	offbase := int64(blockIdx) << r.BlockBits
	h := crc32.NewIEEE()

	var bodyLength int

	// The block never be uploaded.
	if blockFirstPut(prog) {
		bodyLength = getBodyLength(chunkSize, blkSize)
		body1 := io.NewSectionReader(f, offbase, int64(bodyLength))
		body := io.TeeReader(body1, h)

		ret, code, err = r.mkBlock(blkSize, body, bodyLength)
		if err != nil {
			return
		}

		if ret.Crc32 != h.Sum32() {
			return
		}

		prog.Ctx = ret.Ctx
		prog.Offset = bodyLength
		prog.RestSize = blkSize - bodyLength

		notify(blockIdx, prog)

	} else if prog.Offset+prog.RestSize != blkSize {
		code, err = 400, errors.New("Invalid args when doing ResumableBlockPut.")
		return
	}

	for prog.RestSize > 0 {
		if chunkSize < prog.RestSize {
			bodyLength = chunkSize
		} else {
			bodyLength = prog.RestSize
		}

		retry := retryTimes
	lzRetry:
		body1 := io.NewSectionReader(f, offbase+int64(prog.Offset), int64(bodyLength))
		h.Reset()
		body := io.TeeReader(body1, h)
		ret, code, err = r.blockPut(prog.Ctx, prog.Offset, body, bodyLength)

		// put successfully, but need more check should be done.
		if err == nil {
			if ret.Crc32 == h.Sum32() {
				prog.Ctx = ret.Ctx
				prog.Offset += bodyLength
				prog.RestSize -= bodyLength
				notify(blockIdx, prog)
				continue
			} else {
				fmt.Println("ResumableBlockPut invalied checksum : ", offbase, prog.Offset, body)
			}
		} else {
			if code == InvalidCtx {
				fmt.Println("Invalid Context 701!")
				prog.Ctx = ""
				notify(blockIdx, prog)
				break
			}
		}

		if retry > 0 {
			retry--
			goto lzRetry
		}

		break
	}

	return
}

// a helper
func (r Service) BlockCount(fsize int64) int {
	blockMask := int64((1 << r.BlockBits) - 1)
	return int((fsize + blockMask) >> r.BlockBits)
}

func worker(tasks chan func()) {
	for {
		task := <-tasks
		task()
	}
}

func (r Service) Put(
	f io.ReaderAt, fsize int64, checksums []string, progs []BlockProgress,
	blockNotify func(blockIdx int, checksum string),
	chunkNotify func(blockIdx int, prog *BlockProgress)) (code int, err error) {

	blockCnt := r.BlockCount(fsize)
	if len(checksums) != blockCnt || len(progs) != blockCnt {
		code, err = 400, errors.New("up.Service.Put")
		return
	}

	var wg sync.WaitGroup
	wg.Add(blockCnt)
	last := blockCnt - 1
	blockSize := 1 << r.BlockBits
	var failed bool
	for i := 0; i < blockCnt; i++ {
		if checksums[i] == "" {
			blockIdx := i
			blockSize1 := blockSize
			if i == last {
				offbase := int64(blockIdx) << r.BlockBits
				blockSize1 = int(fsize - offbase)
			}
			task := func() {
				defer wg.Done()
				ret, code, err2 := r.resumableBlockput(
					f, blockIdx, blockSize1, r.PutChunkSize, PUT_RETRY_TIMES, &progs[blockIdx], chunkNotify)
				if err2 != nil {
					fmt.Println("ResumableBockPut", blockIdx, "failed", code, err2)
					failed = true
				} else {
					checksums[blockIdx] = ret.Checksum
					blockNotify(blockIdx, ret.Checksum)
				}
				progs[blockIdx].Err = err2
			}
			r.Tasks <- task
		} else {
			wg.Done()
		}
	}

	wg.Wait()
	if failed {
		code, err = 201, errors.New("Function fails")
	} else {
		code = 200
	}
	return
}

func (r Service) Mkfile(
	ret interface{}, cmd, entry string,
	fsize int64, params, callbackParams string, checksums []string) (code int, err error) {
	if callbackParams != "" {
		params += "/params/" + httputil.EncodeURI(callbackParams)
	}

	n := len(checksums)
	body := make([]byte, 20*n)
	for i, checksum := range checksums {
		ret, err2 := base64.URLEncoding.DecodeString(checksum)
		if err2 != nil {
			code, err = 400, errors.New("mkfile error")
			return
		}
		copy(body[i*20:], ret)
	}
	code, err = r.Conn.CallWithEx(
		ret, r.ip+cmd+httputil.EncodeURI(entry)+"/fsize/"+strconv.FormatInt(fsize, 10)+params,
		r.host, "application/octet-stream", bytes.NewReader(body), (int64)(len(body)))
	return
}
