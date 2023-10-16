package hin

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/spf13/viper"
	"io"
)

func FileMD5(file io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func FileUrl(path string) string {
	return viper.GetString("file.domain") + "/" + path
}
