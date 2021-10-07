package public

import (
	"crypto/sha256"
	"fmt"
)

func SaltPassword(salt, password string) string {
	s1 := sha256.New()
	s1.Write([]byte(password))
	// 得到 password 的 sha256   %x 表示 16 进制
	str1 := fmt.Sprintf("%x", s1.Sum(nil))

	s2 := sha256.New()
	s2.Write([]byte(str1 + salt))
	return fmt.Sprintf("%x", s2.Sum(nil))

}
