package util

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var log = logrus.WithField("com", "util")

//Zip 压缩
func Zip(data []byte) (compressed []byte) {
	var b bytes.Buffer
	wr := gzip.NewWriter(&b)
	wr.Write(data)
	wr.Flush()
	wr.Close()
	return b.Bytes()
}

//Unzip 解压缩
func Unzip(compressedData []byte) (data []byte, err error) {
	rd, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Errorf("gzip.NewReader failed. err:%s", err.Error())
		return data, err
	}

	data, err = ioutil.ReadAll(rd)
	if err != nil {
		log.Errorf("ioutil.ReadAll failed. err:%s", err.Error())
		return data, err
	}

	return data, nil
}

//Zlib 压缩
func Zlib(data []byte) (compresseData []byte, err error) {
	var b bytes.Buffer
	wr, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		return compresseData, err
	}

	wr.Write(data)
	wr.Flush()
	wr.Close()
	return b.Bytes(), nil
}

//Unzlib 解压缩
func Unzlib(compresseData []byte) (uncompresseData []byte, err error) {
	rd, err := zlib.NewReader(bytes.NewReader(compresseData))
	if err != nil {
		log.Errorf("zlib.NewReader failed. err:%s", err.Error())
		return uncompresseData, err
	}

	uncompresseData, err = ioutil.ReadAll(rd)
	if err != nil {
		log.Errorf("ioutil.ReadAll failed. err:%s", err.Error())
		return uncompresseData, err
	}

	return uncompresseData, nil
}
