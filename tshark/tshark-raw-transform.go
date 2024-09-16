package main

import (
    "fmt"
    "strings"
    "regexp"
    "strconv"
    "bufio"
    "os"
)

func main() {
	prefix := "_raw\":\""
	postfix := "\""
    search := regexp.MustCompile(prefix + "([0-9a-fA-F]{2})+" + postfix)
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 32*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		line = search.ReplaceAllStringFunc(line, func(match string) string {
			var sb strings.Builder
			length := len(match)-len(postfix)
			for i := len(prefix); i < length; i+=2 {
				hex := match[i : (i+2)]
				value, _ := strconv.ParseInt(hex, 16, 8)
				if 0x20 <= value && value <= 0x7e {
					if value == 0x22 || value == 0x5c { //escape quote and backslash
						sb.WriteString("\\")
					} 
					sb.WriteRune(rune(value))
				} else {
					sb.WriteString("\\u00")
					sb.WriteString(hex)
				}
			}
			return prefix + sb.String() + postfix
		})
		fmt.Println(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}