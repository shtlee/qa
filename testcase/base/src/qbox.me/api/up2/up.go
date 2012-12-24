package up2

import (
	"io"
	"errors"
	"sync"
	"strings"
	"strconv"
	"net/http"
	"hash/crc32"
	"qbox.us/rpc"
	"qbox.me/httputil"
	"qbox.me/errcode"
)


type Service struct {
	host, ip string
	BlockBits uint
	RPutChunkSize, RPutRetryTimes int
	Conn *httputil.Client
}


func New(host, ip string, blockbits uint, chunksize, retryTimes int, t http.RoundTripper) (s *Service, err error) {

	if t == nil {
		t = http.DefaultTransport
	}
	client := &http.Client{Transport: t}
	s = &Service{host, ip, blockbits, chunksize, retryTimes, &httputil.Client{client}}
	return
}


type BlockputProgress struct {
	Ctx string  `json:"ctx"`
	Checksum string `json:"checksum"`
	Crc32 uint32 `json:"crc32"`
	Offset int64 `json:"offset"`
}

type RPtask struct {
	*Service
	EntryURI string
	Type string
	Size int64
	Customer, Meta string
	CallbackParams string
	Body io.ReaderAt
	Progress []BlockputProgress
	ChunkNotify, BlockNotify func(blockIdx int, prog *BlockputProgress)
}

// Create a new resumable put task
func (s *Service) NewRPtask(
	entryURI, mimeType string, customer, meta, params string,
	r io.ReaderAt, size int64, progs []BlockputProgress) (t *RPtask) {

	blockcnt := int((size + (1 << s.BlockBits) - 1) >> s.BlockBits)
	if progs == nil {
		progs = make([]BlockputProgress, blockcnt)
	}
	return &RPtask{s, entryURI, mimeType, size, customer, meta, params, r, progs, nil, nil}
}


// Running the resumable put task
func (t *RPtask) Run(taskQsize, threadSize int,
	chunkNotify, blockNotify func(blockIdx int, prog *BlockputProgress)) (code int, err error) {

	var (
		wg sync.WaitGroup
		failed bool
	)
	worker := func(tasks chan func()) {
		for {
			task := <- tasks
			task()
		}
	}
	blockcnt := len(t.Progress)
	t.ChunkNotify = chunkNotify
	t.BlockNotify = blockNotify

	if taskQsize == 0 {
		taskQsize = blockcnt
	}
	if threadSize == 0 {
		threadSize = blockcnt
	}

	tasks := make(chan func(), taskQsize)
	for i := 0; i < threadSize; i++ {
		go worker(tasks)
	}

	wg.Add(blockcnt)
	for i := 0; i < blockcnt; i++ {
		blkIdx := i
		task := func() {
			defer wg.Done()
			code, err = t.PutBlock(blkIdx)
			if err != nil {
				failed = true
			}
		}
		tasks <- task
	}
	wg.Wait()

	if failed {
		return 400, err
	}

	return t.Mkfile()
}


func (t *RPtask) PutBlock(blockIdx int) (code int, err error) {
	var (
		url string
		restsize, blocksize int64
	)

	h := crc32.NewIEEE()
	prog := &t.Progress[blockIdx]
	offbase := int64(blockIdx << t.BlockBits)

	// blocksize
	if blockIdx == len(t.Progress) - 1 {
		blocksize = t.Size - offbase
	} else {
		blocksize = int64(1 << t.BlockBits)
	}

	initProg := func(p *BlockputProgress) {
		p.Offset = 0
		p.Ctx = ""
		p.Crc32 = 0
		p.Checksum = ""
		restsize = blocksize
	}

	if prog.Ctx == "" {
		initProg(prog)
	}

	for restsize > 0 {
		bdlen := int64(t.RPutChunkSize)
		if bdlen > restsize {
			bdlen = restsize
		}
		retry := t.RPutRetryTimes
	lzRetry:
		h.Reset()
		bd1 := io.NewSectionReader(t.Body, int64(offbase + prog.Offset), int64(bdlen))
		bd := io.TeeReader(bd1, h)
		if prog.Ctx == "" {
			url = t.ip + "/mkblk/" + strconv.FormatInt(blocksize, 10)
		} else {
			url = t.ip + "/bput/" + prog.Ctx + "/" + strconv.FormatInt(prog.Offset, 10)
		}
		code, err = t.Conn.CallWithEx(prog, url, t.host, "application/octet-stream", bd, bdlen)
		if err == nil {
			if prog.Crc32 == h.Sum32() {
				restsize = blocksize - prog.Offset
				if t.ChunkNotify != nil {
					t.ChunkNotify(blockIdx,prog)
				}
				continue
			} else {
				err = errors.New("ResumableBlockPut: Invalid Checksum")
			}
		}
		if code == errcode.InvalidCtx {
			initProg(prog)
			err = errcode.EInvalidCtx
			continue   // retry upload current block
		}
		if retry > 0 {
			retry--
			goto lzRetry
		}
		break
	}
	if t.BlockNotify != nil {
		t.BlockNotify(blockIdx,prog)
	}
	return
}





func (t *RPtask) Mkfile() (code int, err error) {
	var (
		ctx string
	)
	for k,p := range t.Progress {
		if k == len(t.Progress) - 1 {
			ctx += p.Ctx
		} else {
			ctx += p.Ctx + ","
		}
	}
	bd := []byte(ctx)
	url := t.ip + "/rs-mkfile/" + rpc.EncodeURI(t.EntryURI)
	url += "/fsize/" + strconv.FormatInt(t.Size, 10)
	if t.Meta != "" {
		url += "/meta/" + rpc.EncodeURI(t.Meta)
	}
	if t.Customer != "" {
		url += "/customer/" + t.Customer
	}
	if t.CallbackParams != "" {
		url += "/params/" + t.CallbackParams
	}
	code, err = t.Conn.CallWithEx(nil, url, t.host, "", strings.NewReader(string(bd)), int64(len(bd)))
	return
}

func (s *Service) Put(
	entryURI, mimeType string, customer, meta, params string,
	body io.ReaderAt, bodyLength int64,
	progs []BlockputProgress,
	chunkNotify, blockNotify func(blockIdx int, prog *BlockputProgress)) (code int, err error) {

	t1 := s.NewRPtask(entryURI, mimeType, customer, meta, params, body, bodyLength, progs)
	return t1.Run(0, 0, chunkNotify, blockNotify)
}
