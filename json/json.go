package json

import (
	"encoding/json"
	"log"
)

func Encode(in interface{}) (res string) {
	if in == nil {
		return ""
	}
	bts, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		log.Panicf("Json encode error: %v. In: %v", err, in)
	}
	return string(bts)
}

func Decode(in string) (res interface{}) {
	if in == "" {
		return nil
	}

	err := json.Unmarshal([]byte(in), &res)
	if err != nil {
		log.Panicf("Json decode error: %v. In: %v", err, in)
	}
	return res
}
