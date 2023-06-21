package utils

import "encoding/json"

func Debug(obj any) string {
	raw, _ := json.MarshalIndent(obj, "", "\t")
	return string(raw)
}
