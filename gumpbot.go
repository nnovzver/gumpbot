package main

import "net/http"
import "fmt"
import "log"
import "io"
import "os"
import "time"
import "encoding/json"
import "strconv"
import "io/ioutil"
import "flag"
import "encoding/binary"

const (
	secret_token_file string = "secret_token"
	msg_offset_file   string = "msg_offset"
)

func readSecreteToken() string {
	token, err := ioutil.ReadFile(secret_token_file)
	if err != nil {
		panic(err)
	}

	return string(token)
}

func readInt64File(fname string) int64 {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal("os.OpenFile", err)
	}
	var offset int64
	// ignore error offset = 0
	binary.Read(f, binary.LittleEndian, &offset)
	f.Close()
	return offset
}

func writeInt64File(fname string, offset int64) {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal("os.OpenFile", err)
	}
	// ignore irrelevant error
	binary.Write(f, binary.LittleEndian, offset)
	f.Close()
}

var apiPrefix string = "https://api.telegram.org/bot" + readSecreteToken()

type Response struct {
	ok     bool
	result []map[string]interface{}
}

func main() {
	var dumpFlag = flag.Bool("d", false, "dump json response into file json_dump")
	flag.Parse()

	var dumpFile *os.File
	var err error
	if *dumpFlag {
		dumpFile, err = os.OpenFile("json_dump", os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			log.Fatal("os.OpenFile", err)
		}
	}
	defer dumpFile.Close()

	var msg_offset int64 = readInt64File(msg_offset_file)

	for {
		resp, err := http.Get(apiPrefix + "/getUpdates?" + "offset=" + strconv.FormatInt(msg_offset, 10))
		if err != nil {
			log.Fatal("http.Get", err)
		}

		var respReader io.Reader
		if *dumpFlag {
			respReader = io.TeeReader(resp.Body, dumpFile)
		} else {
			respReader = resp.Body
		}

		dec := json.NewDecoder(respReader)
		for {
			fmt.Println("----")
			var v map[string]interface{}
			if err := dec.Decode(&v); err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal("Decode", err)
			}
			for k := range v {
				if k == "ok" {
					fmt.Println("ok =", v[k])
				}
				if k == "result" {
					results := v[k].([]interface{})
					fmt.Println("result =", results)
					for result_index, result := range results {
						fmt.Println()
						fmt.Println("- result =", result_index, "-", result)
						update_id := result.(map[string]interface{})["update_id"].(float64)
						message := result.(map[string]interface{})["message"]
						fmt.Println("update_id =", update_id)
						fmt.Println("message =", message)
						message_text := message.(map[string]interface{})["text"]
						message_chat := message.(map[string]interface{})["chat"]
						message_chat_id := message_chat.(map[string]interface{})["id"].(float64)
						message_date := message.(map[string]interface{})["date"]
						message_id := message.(map[string]interface{})["message_id"]
						message_from := message.(map[string]interface{})["from"]
						fmt.Println("message_text =", message_text)
						fmt.Println("message_chat =", message_chat)
						fmt.Println("message_chat_id =", message_chat_id)
						fmt.Println("message_date =", message_date)
						fmt.Println("message_id =", message_id)
						fmt.Println("message_from =", message_from)

						msg_offset = int64(update_id) + 1

						_, err := http.Get(apiPrefix + "/sendMessage?" +
							"chat_id=" + strconv.FormatInt(int64(message_chat_id), 10) +
							"&text=Hello")
						if err != nil {
							log.Fatal("http.Get", err)
						}
						writeInt64File(msg_offset_file, msg_offset)
					}
				}
			}
		}

		resp.Body.Close()
		time.Sleep(10 * time.Second)
	}
}
