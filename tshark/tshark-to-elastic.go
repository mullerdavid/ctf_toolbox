package main

import (
    "strings"
    "bytes"
    "regexp"
    "strconv"
    "bufio"
    "fmt"
    "os"
	"net/http"
	"io/ioutil"
)

func createHexDecoder() (f func(json string) string) {
	const prefix = "_raw\":\""
	const postfix = "\""
    search := regexp.MustCompile(prefix + "([0-9a-fA-F]{2})+" + postfix)
	return func(json string) string {
		json = search.ReplaceAllStringFunc(json, func(match string) string {
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
		return json
	}
}

func sendBulkToElastic(host string, buf []byte) {
	retries := 3
	client := &http.Client{}
	for 0 <= retries {
		retries -= 1
        req, err := http.NewRequest("POST", host, bytes.NewBuffer(buf))
		if err != nil {
			fmt.Println(err)
            continue
        } 
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
        if err != nil {
			fmt.Println(err)
            continue
        }
		defer res.Body.Close()
		if true {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(body))
		}
		break
    }
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "http://elasticsearch:9200/_bulk")
		fmt.Println("Example:", os.Args[0], "http://elasticsearch:9200/_bulk")
	} else {
		elasticHost := os.Args[1]
	
		const batch = 1024
		decoder := createHexDecoder()
		scanner := bufio.NewScanner(os.Stdin)
		readBuf := make([]byte, 0, 1024*1024)
		scanner.Buffer(readBuf, 32*1024*1024)
		counter := 0
		var writeBuf bytes.Buffer
	
		for scanner.Scan() {
			line := scanner.Text()
			content := !strings.HasPrefix(line, "{\"index\":{\"_index\"")
			if content {
				// decoding fields ending _raw and containing hex data to plaintext
				// unreadable characters are encoded as \u00xx
				line = decoder(line)
				counter ++
			} else {
				// removing _type from indices
				line = strings.Replace(line, ",\"_type\":\"doc\"", "", 1)
			}
			writeBuf.Write([]byte(line))
			writeBuf.WriteRune('\n')
			if content && counter % batch == 0 {
				sendBulkToElastic(elasticHost, writeBuf.Bytes() )
				writeBuf.Reset()
			}
		}
		if counter % batch != 0 {
			sendBulkToElastic(elasticHost, writeBuf.Bytes() )
		}
	
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}
}

//TODO: error handling 