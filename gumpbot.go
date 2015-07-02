package main

import "net/http"
import "log"
import "io"
import "os"
import "time"
import "encoding/json"
import "strconv"
import "io/ioutil"
import "flag"
import "encoding/binary"
import "net/url"

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

var apiSecretToken = readSecreteToken()

func makeApiUrl(cmd string, args url.Values) string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = "api.telegram.org"
	u.Path = "bot" + apiSecretToken + "/" + cmd
	u.RawQuery = args.Encode()
	return u.String()
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
		reqestArgs := url.Values{}
		reqestArgs.Add("offset", strconv.FormatInt(msg_offset, 10))
		resp, err := http.Get(makeApiUrl("getUpdates", reqestArgs))
		if err != nil {
			log.Fatal("http.Get", err)
		}

		var respReader io.Reader
		if *dumpFlag {
			respReader = io.TeeReader(resp.Body, dumpFile)
		} else {
			respReader = resp.Body
		}

		updates, err := UnmarshalResponse(respReader)
		if err != nil {
			log.Fatal("UnmarshalResponse", err)
		}

		for _, u := range updates {
			log.Printf("%#v\n", u)
			msg_offset = int64(u.update_id) + 1

			reqestArgs = url.Values{}
			reqestArgs.Add("chat_id", strconv.FormatInt(u.chat_id, 10))
			reqestArgs.Add("text", "This bot just do nothing for you.\n"+
				"Simply send anything you want and read this message =)\n\n"+
				"You can use any command to get results, it doesn't help ^_____^")
			_, err := http.Get(makeApiUrl("sendMessage", reqestArgs))
			if err != nil {
				log.Fatal("http.Get", err)
			}
			writeInt64File(msg_offset_file, msg_offset)
		}

		resp.Body.Close()
		time.Sleep(10 * time.Second)
	}
}

type UpdatePayload struct {
	update_id int64
	chat_id   int64
	text      string
}

func UnmarshalResponse(respReader io.Reader) ([]UpdatePayload, error) {
	dec := json.NewDecoder(respReader)
	var resp map[string]interface{}
	var err error
	var updates []UpdatePayload
	for {
		err = dec.Decode(&resp)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return nil, err
		}

		for _, u := range resp["result"].([]interface{}) {
			updates = append(updates, UpdatePayload{})
			updates[len(updates)-1].update_id = int64(u.(map[string]interface{})["update_id"].(float64))
			updates[len(updates)-1].text = u.(map[string]interface{})["message"].(map[string]interface{})["text"].(string)
			updates[len(updates)-1].chat_id = int64(u.(map[string]interface{})["message"].(map[string]interface{})["chat"].(map[string]interface{})["id"].(float64))
		}
	}
	return updates, err
}
