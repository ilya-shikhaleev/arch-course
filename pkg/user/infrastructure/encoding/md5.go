package encoding

import (
	"crypto/md5"
	"fmt"

	"github.com/ilya-shikhaleev/arch-course/pkg/user/app/user"
)

func MD5Encoder() user.EncoderFunc {
	return func(s string) string {
		data := []byte(s)
		return fmt.Sprintf("%x", md5.Sum(data))
	}
}
