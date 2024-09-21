package main

import (
    "strings"
    "bytes"
    "regexp"
    "strconv"
    "bufio"
    "fmt"
    "os"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

const batch = 1024*1024*15 // 15MiB

type JSONMap = map[string]interface{}
type JSONArray = []interface{}

func createHexDecoder() (f func(json string) string) {
    search := regexp.MustCompile("([0-9a-fA-F]{2})+")
	return func(json string) string {
		json = search.ReplaceAllStringFunc(json, func(match string) string {
			var sb strings.Builder
			length := len(match)
			for i := 0; i < length; i+=2 {
				hex := match[i : (i+2)]
				value, _ := strconv.ParseInt(hex, 16, 9)
				sb.WriteRune(rune(value))
			}
			return sb.String()
		})
		return json
	}
}

var hexDecode func(string) string = createHexDecoder() 

func hexDecodeArray(arr JSONArray) JSONArray {
	for k, v := range arr {
        arr[k] = hexDecode(v.(string))
    }
	return arr
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
		if false {
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

func read(r *bufio.Reader) ([]byte, error) {
    var (
        isPrefix = true
        err      error
        line, ln []byte
    )

    for isPrefix && err == nil {
        line, isPrefix, err = r.ReadLine()
		if len(line) < batch*2 { //read full line, but don't copy if too big
        	ln = append(ln, line...)
		}
    }

    return ln, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Reads tshark output from stdin and transforms and sends to elasticsearch bulk endpoint.")
		fmt.Println("Usage:", os.Args[0], "http://elasticsearch:9200/_bulk")
		fmt.Println("Example: tshark -T ek -J \"http tcp udp ip\" -x -r ./dump.pcap |", os.Args[0], "http://localhost:9200/packets_template/_bulk")
	} else {
		elasticHost := os.Args[1]

		reader := bufio.NewReader(os.Stdin)
		counter := 0
		var writeBuf bytes.Buffer

		for {
			line, err := read(reader)
			if err != nil {
				if err != io.EOF {
				}
				break
				fmt.Println("Error:", err)
			}
	
			var jsonmap JSONMap
			var exists bool
			var node string
			var nodeArr JSONArray
			if len(line) < batch*2 {
				_ = json.Unmarshal([]byte(line), &jsonmap)
			} else {
				fmt.Println("Skipping record, too big")
				_ = json.Unmarshal([]byte("{\"too_big\":true}"), &jsonmap)
			}
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
						node_udp["udp_udp_payload_raw"] = hexDecode(node)
					}
				}
				node_tcp, tcp := node_layers["tcp"].(JSONMap)
				if tcp {
					node, exists = node_tcp["tcp_tcp_payload_raw"].(string)
					if (exists) {
						node_tcp["tcp_tcp_payload_raw"] = hexDecode(node)
					}
				}
				node_http, http := node_layers["http"].(JSONMap)
				if http {
					node, exists = node_http["http_http_data_raw"].(string)
					if (exists) {
						node_http["http_http_data_raw"] = hexDecode(node)
					}
					nodeArr, exists = node_http["http_http_request_line_raw"].(JSONArray)
					if (exists) {
						node_http["http_http_request_line_raw"] = hexDecodeArray(nodeArr)
					}
					nodeArr, exists = node_http["http_http_response_line_raw"].(JSONArray)
					if (exists) {
						node_http["http_http_response_line_raw"] = hexDecodeArray(nodeArr)
					}
					if false {
						fmt.Println(nodeArr)
					}
				}
				counter ++
			}
			buff, _ := json.Marshal(jsonmap)
			writeBuf.Write(buff)
			writeBuf.WriteRune('\n')
			if data && (batch < writeBuf.Len() ) {
				sendBulkToElastic(elasticHost, writeBuf.Bytes() )
				fmt.Println("Written", counter, "records overall,", writeBuf.Len(), "bytes this batch")
				writeBuf.Reset()
			}
		}
		if 0 < writeBuf.Len() {
			sendBulkToElastic(elasticHost, writeBuf.Bytes() )
			fmt.Println("Written ", counter, "records overall,", writeBuf.Len(), "bytes this batch")
		}
	}
}