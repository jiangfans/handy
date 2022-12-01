package utils

import (
	"bytes"
	"compress/zlib"
	log "github.com/sirupsen/logrus"
	"io"
)

// ZlibCompress 进行zlib压缩
func ZlibCompress(src []byte) ([]byte, error) {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, err := w.Write(src)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	err = w.Close()
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return in.Bytes(), nil
}

// ZlibUnCompress 进行zlib解压缩
func ZlibUnCompress(compressSrc []byte) ([]byte, error) {
	bs := bytes.NewReader(compressSrc)
	r, err := zlib.NewReader(bs)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, r)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return out.Bytes(), nil
}
