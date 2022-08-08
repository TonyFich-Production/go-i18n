package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"

	"github.com/nicksnyder/go-i18n/v3/i18n"
	"github.com/nicksnyder/go-i18n/v3/internal/plural"
)

func writeFile(outdir, label string, langTag language.Tag, format string, messageTemplates map[i18n.MessageID]*i18n.MessageTemplate, sourceLanguage bool) (path string, content []byte, err error) {
	v := marshalValue(convertTemplates(messageTemplates), sourceLanguage)
	content, err = marshal(v, format)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal %s strings to %s: %s", langTag, format, err)
	}
	path = filepath.Join(outdir, fmt.Sprintf("%s.%s.%s", label, langTag, format))
	return
}

func convertTemplates(in map[i18n.MessageID]*i18n.MessageTemplate) map[string]*i18n.MessageTemplate {
	out := make(map[string]*i18n.MessageTemplate, len(in))
	for k, v := range in {
		out[string(k)] = v
	}

	log.Printf("in:  %+v", in)
	log.Printf("out: %+v", out)

	return out
	// return *(*map[string]*i18n.MessageTemplate)(unsafe.Pointer(&in))
}

func marshalValue(messageTemplates map[string]*i18n.MessageTemplate, sourceLanguage bool) interface{} {
	v := make(map[string]interface{}, len(messageTemplates))
	for id, template := range messageTemplates {
		if other := template.PluralTemplates[plural.Other]; sourceLanguage && len(template.PluralTemplates) == 1 &&
			other != nil && template.Description == "" && template.LeftDelim == "" && template.RightDelim == "" {
			v[id] = other.Src
		} else {
			m := map[string]string{}
			if template.Description != "" {
				m["description"] = template.Description
			}
			if !sourceLanguage {
				m["hash"] = template.Hash
			}
			for pluralForm, template := range template.PluralTemplates {
				m[string(pluralForm)] = template.Src
			}
			v[id] = m
		}
	}
	return v
}

func marshal(v interface{}, format string) ([]byte, error) {
	switch format {
	case "json":
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		err := enc.Encode(v)
		return buf.Bytes(), err
	case "toml":
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		enc.Indent = ""
		err := enc.Encode(v)
		return buf.Bytes(), err
	case "yaml":
		return yaml.Marshal(v)
	}
	return nil, fmt.Errorf("unsupported format: %s", format)
}
