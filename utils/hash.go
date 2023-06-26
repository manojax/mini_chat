package utils

import (
	"crypto/md5"
	"fmt"
)

func HashMD5(s string) string {
	hash := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)
}
