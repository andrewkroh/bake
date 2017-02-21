package common

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonFindGitProjectRoot(t *testing.T) {
	root, err := FindGitProjectRoot()
	if err != nil {
		t.Fatal(err)
	}

	expectedRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectedRoot, root)
}
