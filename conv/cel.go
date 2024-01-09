package conv

import "encoding/json"

func AnyToMapStringAny(v any) (map[string]any, error) {
	var jsonObj map[string]any
	b, err := json.Marshal(v)
	if err != nil {
		return jsonObj, err
	}
	err = json.Unmarshal(b, &jsonObj)
	return jsonObj, err
}

func AnyToListMapStringAny(v any) ([]map[string]any, error) {
	var jsonObj []map[string]any
	b, err := json.Marshal(v)
	if err != nil {
		return jsonObj, err
	}
	err = json.Unmarshal(b, &jsonObj)
	return jsonObj, err
}
