package xserver

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

// 高效率生成随机数方法
type randomStringGenerator struct {
	// 获取随机数不是并发安全的
	l         sync.Mutex
	src       rand.Source
	length    int
	remain    int
	cache     int64
	cnt       int
	needReset bool
}

func NewRandomStringGenerator(length int) *randomStringGenerator {
	var r randomStringGenerator
	r.src = rand.NewSource(time.Now().UnixNano())
	r.length = length
	return &r
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// 重置
// rand.Int63 目前碰到的问题：
// 获取一定数量的随机数之后
// 后面的随机数都会变成0
// 暂时没有查到问题所在
// 增加一个重置的方法来重置随机数
func (r *randomStringGenerator) Reset() {
	r.l.Lock()
	r.src = rand.NewSource(time.Now().UnixNano())
	r.needReset = false
	r.l.Unlock()
}

func (r *randomStringGenerator) RandStringBytesMaskImprSrcUnsafe() string {
	r.l.Lock()
	defer r.l.Unlock()
	b := make([]byte, r.length)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i := r.length - 1; i >= 0; {
		if r.remain == 0 {
			r.cache, r.remain = r.src.Int63(), letterIdxMax
			r.cnt++
		}

		if idx := int(r.cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}

		r.cache >>= letterIdxBits
		r.remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
