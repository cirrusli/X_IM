package fuzz

import (
	"testing"
	"unicode/utf8"
)

// 该函数对于ASCII码里的字符组成的字符串是可以正确反转
// 但是对于非ASCII码里的字符，如果简单按照字节进行反转，得到的可能是一个非法的字符串
func FuzzReverse(f *testing.F) {
	//seed corpus
	strSlice := []string{"abc", "bb"}

	for _, v := range strSlice {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, str string) {
		revStr1 := reverse(str)
		revStr2 := reverse(revStr1)
		if str != revStr2 {
			t.Errorf("fuzz test failed. str:%s, rev_str1:%s, rev_str2:%s", str, revStr1, revStr2)
		}
		if utf8.ValidString(str) && !utf8.ValidString(revStr1) {
			t.Errorf("reverse result is not utf8. str:%s, len: %d, rev_str1:%s", str, len(str), revStr1)
		}
	})
}
func reverse(s string) string {
	bs := []byte(s)
	length := len(bs)
	for i := 0; i < length/2; i++ {
		bs[i], bs[length-i-1] = bs[length-i-1], bs[i]
	}
	return string(bs)
}
