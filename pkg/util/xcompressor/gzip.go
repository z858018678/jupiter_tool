package xcompressor

import (
	"bytes"
	"compress/gzip"
	"io"
)

type gzipCompressor struct {
	compressLevel int
}

// level between -2 ~ 9
func (c *gzipCompressor) Compress(in []byte) ([]byte, error) {
	var b bytes.Buffer
	var w, err = gzip.NewWriterLevel(&b, c.compressLevel)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	_, err = w.Write(in)
	if err != nil {
		return nil, err
	}
	w.Flush()

	return b.Bytes(), nil
}

func (c *gzipCompressor) Uncompress(in []byte) ([]byte, error) {
	var reader = bytes.NewReader(in)

	var r, err = gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	io.Copy(&b, r)
	r.Close()

	return b.Bytes(), nil
}

func NewGzipCompressor(compressLevel int) XCompressor {
	return &gzipCompressor{compressLevel: compressLevel}
}
