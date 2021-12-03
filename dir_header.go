package main

// Reference link: https://chromium.googlesource.com/chromium/src/+/refs/heads/main/net/base/dir_header.html

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"text/template"
)

//go:embed dir_header.html
var dirHtmlTemplate string

type DirHeader struct {
	TextDirection string `json:"textdirection"`
	Language      string `json:"language"`

	Header        string `json:"header"`
	ParentDirText string `json:"parentDirText"`

	HeaderName         string `json:"headerName"`
	HeaderSize         string `json:"headerSize"`
	HeaderDateModified string `json:"headerDateModified"`
}

type Row struct {
	Name               string `json:"name"`
	Url                string `json:"url"`
	IsDir              bool   `json:"isdir"`
	Size               int64  `json:"size"`
	SizeString         string `json:"size_string"`
	DateModified       int64  `json:"date_modified"`
	DateModifiedString string `json:"date_modified_string"`
}

func RenderMap(templateStr string, data map[string]interface{}) ([]byte, error) {
	t := template.Must(template.New("t").Parse(templateStr))
	var b bytes.Buffer
	err := t.Execute(&b, data)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func DirHtml(header *DirHeader, rows []Row) ([]byte, error) {
	headerData, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(headerData, &data)
	if err != nil {
		return nil, err
	}
	rowsData, err := json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	data["rows"] = string(rowsData)
	return RenderMap(dirHtmlTemplate, data)
}
