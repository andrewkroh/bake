package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoticeTemplate(t *testing.T) {
	params := noticeParams{
		BeatName:           "Elastic Beats",
		Copyright:          "Elasticsearch BV",
		CopyrightYearStart: 2014,
		CopyrightYearEnd:   2017,
		Projects: []*projectInfo{
			{"github.com/elastic/go-lumber", "Apache License"},
		},
	}

	buf := new(bytes.Buffer)
	if err := noticeTemplate.Execute(buf, params); err != nil {
		t.Fatal(err)
	}

	expect := `Elastic Beats
Copyright 2014-2017

This product includes software developed by The Apache Software
Foundation (http://www.apache.org/).

==========================================================================
Third party libraries used by the Elastic Beats:
==========================================================================


--------------------------------------------------------------------
github.com/elastic/go-lumber
--------------------------------------------------------------------
Apache License
`

	assert.Equal(t, expect, buf.String())
}

func TestNoticeGetLibraryName(t *testing.T) {
	path := "../../../vendor/github.com/StackExchange/wmi/LICENSE"

	// Replace slashes with OS specific separator.
	strings.Replace(path, "/", string(filepath.Separator), -1)

	name := getLibraryName(path)
	assert.Equal(t, "github.com/StackExchange/wmi", name)
}
