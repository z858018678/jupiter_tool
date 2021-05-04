// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xrand

import (
	"math/rand"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// Charsets
const (
	// Uppercase ...
	Uppercase string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Lowercase ...
	Lowercase = "abcdefghipqrstuvwxyz"
	// Alphabetic ...
	Alphabetic = Uppercase + Lowercase
	// Numeric ...
	Numeric = "0123456789"
	// Alphanumeric ...
	Alphanumeric = Alphabetic + Numeric
	// Symbols ...
	Symbols = "`" + `~!@#$%^&*()-_+={}[]|\;:"<>,./?`
	// Hex ...
	Hex = Numeric + "abcdef"
)

// String 返回随机字符串，通常用于测试mock数据
func String(length uint8, charsets ...string) string {
	charset := strings.Join(charsets, "")
	if charset == "" {
		charset = Alphanumeric
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Int63()%int64(len(charset))]
	}
	return string(b)
}

// 生成随机字符串
// 高效率生成随机数方法
type randomStringGenerator struct {
	// 获取随机数不是并发安全的
	l         sync.Mutex
	src       rand.Source
	trunk     string
	length    int
	remain    int
	cache     int
	cnt       int
	needReset bool
	// to represent a letter index
	letterIdxBits int
	// All 1-bits, as many as letterIdxBits
	letterIdxMask int
	// # of letter indices fitting in 63 bits
	letterIdxMax int
}

func NewRandomStringGenerator(length int, trunk string) *randomStringGenerator {
	var r randomStringGenerator
	r.trunk = trunk
	r.src = rand.NewSource(time.Now().UnixNano())
	r.length = length

	var i, l, n = 1, len(r.trunk), 0

	for l >= i {
		i <<= 1
		n++
		r.letterIdxBits = n
	}

	r.letterIdxMask = 1<<r.letterIdxBits - 1
	r.letterIdxMax = 63 / r.letterIdxBits
	return &r
}

// RandStringBytesMaskImprSrcUnsafe
func (r *randomStringGenerator) New() string {
	// 这里必须加锁，不然并发不安全
	r.l.Lock()
	defer r.l.Unlock()
	b := make([]byte, r.length)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i := r.length - 1; i >= 0; {
		if r.remain == 0 {
			r.cache, r.remain = int(r.src.Int63()), r.letterIdxMax
			r.cnt++
		}

		if idx := r.cache & r.letterIdxMask; idx < len(r.trunk) {
			b[i] = r.trunk[idx]
			i--
		}

		r.cache >>= r.letterIdxBits
		r.remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
