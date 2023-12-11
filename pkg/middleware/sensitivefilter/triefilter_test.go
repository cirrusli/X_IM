package sensitivefilter

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestTrieFilter(t *testing.T) {

}

func TestFilters(t *testing.T) {
	runePrint()

	sensitiveWords := []string{
		"傻逼",
		"傻叉",
		"垃圾",
		"妈的",
		"sb",
	}
	matchContents := []string{
		"你是一个大傻&逼，大傻 叉",
		"你是傻☺叉",
		"shabi东西",
		"他made东西",
		"什么垃圾打野，傻逼一样，叫你来打团不来，SB",
		"正常的内容☺",
	}

	pinyinContents := Hans2Pinyin(sensitiveWords)
	fmt.Println("pinyin: ", pinyinContents)

	forceFilter(sensitiveWords, matchContents)

	regFilter(sensitiveWords, matchContents)

	trieFilter(sensitiveWords, matchContents)
}

func runePrint() {
	fmt.Println("a -> ", rune('a'))
	fmt.Println("A -> ", rune('A'))
	fmt.Println("你 -> ", rune('你'))
	fmt.Println("我 -> ", rune('我'))
	fmt.Println("大家 -> ", []rune("大家"))
}

func forceFilter(sensitiveWords []string, matchContents []string) {
	fmt.Println("\n--------- 普通暴力匹配敏感词 ---------")
	for _, text := range matchContents {
		srcText := text
		for _, word := range sensitiveWords {
			replaceChar := ""

			for i, wordLen := 0, len([]rune(word)); i < wordLen; i++ {
				// 根据敏感词的长度构造和谐字符
				replaceChar += "*"
			}

			text = strings.Replace(text, word, replaceChar, -1)
		}
		fmt.Println("srcText     -> ", srcText)
		fmt.Println("replaceText -> ", text)
		fmt.Println()
	}

}

func regFilter(sensitiveWords []string, matchContents []string) {
	fmt.Println("\n--------- 正则匹配敏感词 ---------")
	banWords := make([]string, 0) // 收集匹配到的敏感词

	// 构造正则匹配字符
	regStr := strings.Join(sensitiveWords, "|")
	wordReg := regexp.MustCompile(regStr)
	println("regStr -> ", regStr)

	for _, text := range matchContents {
		textBytes := wordReg.ReplaceAllFunc([]byte(text), func(bytes []byte) []byte {
			banWords = append(banWords, string(bytes))
			textRunes := []rune(string(bytes))
			replaceBytes := make([]byte, 0)
			for i, runeLen := 0, len(textRunes); i < runeLen; i++ {
				replaceBytes = append(replaceBytes, byte('*'))
			}
			return replaceBytes
		})
		fmt.Println("srcText        -> ", text)
		fmt.Println("replaceText    -> ", string(textBytes))
		fmt.Println("sensitiveWords -> ", banWords)
		fmt.Println()
	}
}

func trieFilter(sensitiveWords []string, matchContents []string) {
	fmt.Println("\n--------- 前缀树匹配敏感词 ---------")
	// 汉字转拼音
	pinyinContents := Hans2Pinyin(sensitiveWords)
	trie := NewTrie()
	trie.AddWords(sensitiveWords)
	trie.AddWords(pinyinContents) // 添加拼音敏感词

	//trieFilter.AddWords(pinyinContents)
	//for _, content := range contents {
	//	trieFilter.AddWord(content)
	//}

	for _, srcText := range matchContents {
		matchSensitiveWords, replaceText := trie.Match(srcText)
		fmt.Println("srcText        -> ", srcText)
		fmt.Println("replaceText    -> ", replaceText)
		fmt.Println("sensitiveWords -> ", matchSensitiveWords)
		fmt.Println()
	}

	// 动态添加
	trie.AddWord("牛大大")
	content := "今天，牛大大挑战灰大大"
	matchSensitiveWords, replaceText := trie.Match(content)
	fmt.Println("srcText        -> ", content)
	fmt.Println("replaceText    -> ", replaceText)
	fmt.Println("sensitiveWords -> ", matchSensitiveWords)
}
