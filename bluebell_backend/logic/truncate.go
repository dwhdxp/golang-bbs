package logic

import (
	"unicode"
	"unicode/utf8"
)

// 向redis中缓存帖子内容时截断字符串，优化查询和减少内存占用

// TruncateByWords 根据最大单词数截断字符串，并添加省略号
func TruncateByWords(s string, maxWords int) string {
	processedWords := 0  // 记录已处理的单词数
	wordStarted := false // 标记当前是否在一个单词内
	for i := 0; i < len(s); {
		// 解码字符串中的每个字符（包括多字节字符）
		r, width := utf8.DecodeRuneInString(s[i:])
		if !isSeparator(r) { // 在一个单词内
			i += width
			wordStarted = true
			continue
		}

		// 处理分隔符，根据上一个处理的是分隔符还是单词
		if !wordStarted { // 上一个处理的是分隔符
			i += width
			continue
		}

		wordStarted = false
		processedWords++
		if processedWords == maxWords {
			const ending = "..."
			if (i + len(ending)) >= len(s) {
				// Source string ending is shorter than "..."
				return s
			}

			return s[:i] + ending
		}

		i += width
	}

	return s
}

// isSeparator 判断一个字符是否为分隔符
func isSeparator(r rune) bool {
	// ASCII字符处理
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}
		return true
	}
	// Unicode字符处理
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	return unicode.IsSpace(r) // 将空格认为是分隔符
}
