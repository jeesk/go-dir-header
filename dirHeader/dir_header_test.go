package dirHeader

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	header := DirHeader{
		TextDirection: "ltr",
		Language:      "en",

		Header:        "Index of /bin/",
		ParentDirText: "[parent directory]",

		HeaderName:         "Name",
		HeaderSize:         "Size",
		HeaderDateModified: "Date Modified",
	}
	rows := []Row{
		{
			Name:               "busybox",
			Url:                "/busybox",
			IsDir:              false,
			Size:               1241141,
			SizeString:         "1.1kb",
			DateModified:       12124121,
			DateModifiedString: "2011/1/4",
		},
		{
			Name:               "lib",
			Url:                "/lib",
			IsDir:              true,
			Size:               2048,
			SizeString:         "2kb",
			DateModified:       12144121,
			DateModifiedString: "2011/1/5",
		},
	}
	html, err := DirHtml(&header, rows)
	assert.Nil(t, err)
	err = ioutil.WriteFile("/tmp/a.html", html, 0o644)
	assert.Nil(t, err)
}
