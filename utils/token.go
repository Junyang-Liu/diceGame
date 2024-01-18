package utils

import (
	"crypto/md5"
	"fmt"
)

func GenToken(path, t, secret string) string {
	source := []byte(fmt.Sprintf("%s%s%s", path, t, secret))
	md5Byte := md5.Sum(source)
	return fmt.Sprintf("%x", md5Byte[:])
}
