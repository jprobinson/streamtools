package blocks

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func PostValue(b *Block) {

	type KeyMapping struct {
		MsgKey   string
		QueryKey string
	}

	type postRule struct {
		Keymapping []KeyMapping
		Endpoint   string
	}

	var rule *postRule

	// TODO check the endpoint for happiness
	for {
		select {
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case msg := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(msg, &postRule{Keymapping:[]KeyMapping{KeyMapping{}}})
			} else {
				marshal(msg, rule)
			}
		case msg := <-b.Routes["set_rule"]:
			if rule == nil {
				rule = &postRule{}
			}
			unmarshal(msg, rule)

		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			body := make(map[string]interface{})
			for _, keymap := range rule.Keymapping {
				keys := strings.Split(keymap.MsgKey, ".")
				value, err := Get(msg, keys...)
				if err != nil {
					log.Println(err.Error())
				} else {
					Set(body, keymap.QueryKey, value)
				}
			}

			// TODO maybe check the response ?
			postBody, err := json.Marshal(body)
			if err != nil {
				log.Fatal(err.Error())
			}

			// TODO the content-type here is heavily borked but we're using a hack
			resp, err := http.Post(rule.Endpoint, "application/x-www-form-urlencoded", bytes.NewReader(postBody))
			if err != nil {
				log.Println(err.Error())
			}
			defer resp.Body.Close()
		}
	}
}