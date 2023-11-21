package middleware

type Filter interface {
	FilterSpecialChar(text string) string
	AddWord(sensitiveWord string)
	AddWords(sensitiveWords []string)
	Match(text string) (sensitiveWords []string, replaceText string)
}
