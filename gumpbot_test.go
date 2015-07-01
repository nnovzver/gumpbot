package main

import "testing"
import "bytes"
import "reflect"

var byteResponse = []byte(`
{"ok":true,
 "result":[{"update_id":521777393,
            "message":{"message_id":53,
                       "from":{"id":12341599,
                               "first_name":"Foo",
                               "last_name":"Barov",
                               "username":"foobarov"},
                       "chat":{"id":12341599,
                               "first_name":"Foo",
                               "last_name":"Barov",
                               "username":"foobarov"},
                       "date":1435680065,
                       "text":"\/help"}},
           {"update_id":521595394,
            "message":{"message_id":54,
                       "from":{"id":12341599,
                               "first_name":"Foo",
                               "last_name":"Barov",
                               "username":"foobarov"},
                       "chat":{"id":12341599,
                               "first_name":"Foo",
                               "last_name":"Barov",
                               "username":"foobarov"},
                       "date":1435680068,
                       "text":"\/start"}}]}`)

func TestUnmarshalResponse(t *testing.T) {
	jreader := bytes.NewReader(byteResponse)
	resp, err := UnmarshalResponse(jreader)
	if err != nil {
		t.Error("UnmarshalResponse", err)
	}

	if !reflect.DeepEqual(resp[0], UpdatePayload{update_id: 521777393, chat_id: 12341599, text: `/help`}) {
		t.Error("invalid resp[0]")
	}
	if !reflect.DeepEqual(resp[1], UpdatePayload{update_id: 521595394, chat_id: 12341599, text: `/start`}) {
		t.Error("invalid resp[1]")
	}
}
