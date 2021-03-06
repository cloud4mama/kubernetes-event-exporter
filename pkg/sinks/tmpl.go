package sinks

import (
	"bytes"
	"encoding/json"
	"github.com/Masterminds/sprig"
	"github.com/opsgenie/kubernetes-event-exporter/pkg/kube"
	"text/template"
)

func GetString(event *kube.EnhancedEvent, text string) (string, error) {
	tmpl, err := template.New("template").Funcs(sprig.TxtFuncMap()).Parse(text)
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	// TODO: Should we send event directly or more events?
	err = tmpl.Execute(buf, event)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func convertLayoutTemplate(layout map[string]interface{}, ev *kube.EnhancedEvent) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range layout {
		switch v := value.(type) {
		case string:
			rendered, err := GetString(ev, v)
			if err != nil {
				return nil, err
			}

			result[key] = rendered
		case map[interface{}]interface{}:
			strKeysMap := make(map[string]interface{})
			for k, v := range v {
				// TODO: It's a bit dangerous
				strKeysMap[k.(string)] = v
			}

			res, err := convertLayoutTemplate(strKeysMap, ev)
			if err != nil {
				return nil, err
			}
			result[key] = res

		}
	}
	return result, nil
}

func serializeEventWithLayout(layout map[string]interface{}, ev *kube.EnhancedEvent) ([]byte, error) {
	var toSend []byte
	if layout != nil {
		res, err := convertLayoutTemplate(layout, ev)
		if err != nil {
			return nil, err
		}

		toSend, err = json.Marshal(res)
		if err != nil {
			return nil, err
		}
	} else {
		toSend = ev.ToJSON()
	}
	return toSend, nil
}
