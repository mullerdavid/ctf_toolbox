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
	"encoding/json"
)

type JSONMap = map[string]interface{}
type JSONArray = []interface{}

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
	fmt.Println(string(buf))
	return
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

func unmarshal[T any](data []byte) (*T, error) {
    out := new(T)
    if err := json.Unmarshal(data, out); err != nil {
        return nil, err
    }
    return out, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:", os.Args[0], "http://elasticsearch:9200/_bulk")
		fmt.Println("Example:", os.Args[0], "http://elasticsearch:9200/_bulk")
	} else {
		elasticHost := os.Args[1]

		const batch = 1024
		scanner := bufio.NewScanner(os.Stdin)
		readBuf := make([]byte, 0, 1024*1024)
		scanner.Buffer(readBuf, 32*1024*1024)
		counter := 0
		var writeBuf bytes.Buffer
	
		for scanner.Scan() {
			line := scanner.Text()
			var jsonmap JSONMap
			var exists bool
			var node string
			var nodeArr JSONArray
			_ = json.Unmarshal([]byte(line), &jsonmap)
			node_index, index := jsonmap["index"].(JSONMap)
			if index {
				// removing _type from indices, elastic 8.15 is not using it (and results in error)
				delete(node_index, "_type")
			} 
			node_layers, data := jsonmap["layers"].(JSONMap)
			if data {
				// decoding some fields ending _raw and containing hex data to plaintext
				// unreadable characters are encoded as \u00xx
				node_udp, udp := node_layers["udp"].(JSONMap)
				if udp {
					node, exists = node_udp["udp_udp_payload_raw"].(string)
					if (exists) {
						node_udp["udp_udp_payload_raw"] = "HELLO" + node
					}
				}
				node_tcp, tcp := node_layers["tcp"].(JSONMap)
				if tcp {
					node, exists = node_tcp["tcp_tcp_payload_raw"].(string)
					if (exists) {
						node_tcp["tcp_tcp_payload_raw"] = "WORLD" + node
					}
				}
				node_http, http := node_layers["http"].(JSONMap)
				if http {
					nodeArr, exists = node_http["http_http_request_line_raw"].(JSONArray)
					if (exists) {
						node_http["http_http_request_line_raw"] = JSONArray{"XXXXXXXXXX"}
					}
					nodeArr, exists = node_http["http_http_response_line_raw"].(JSONArray)
					if (exists) {
						node_http["http_http_response_line_raw"] = JSONArray{"ZZZZZZZZZZ"}
					}
					if false {
						fmt.Println(nodeArr)
					}
				}
				counter ++
			}
			// TODO: Encoder.SetEscapeHTML(false)
			buff, _ := json.Marshal(jsonmap)
			writeBuf.Write(buff)
			writeBuf.WriteRune('\n')
			if data && counter % batch == 0 {
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