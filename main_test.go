package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/stretchr/testify/assert"
)

// Test pretty printing of samples in testdata/ by comparing with reference
func TestPrettyPrint(t *testing.T) {
	err := filepath.WalkDir("testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(d.Name()) != ".sample" {
			return nil
		}

		t.Run(d.Name(), func(t *testing.T) {
			f, err := os.Open(path)
			defer f.Close()
			assert.NoErrorf(t, err, "open %q", path)

			var out bytes.Buffer

			runner, err := NewRunner(PrettyStateMachine, f, &out)
			assert.NoError(t, err, "new runner")

			err = runner.Run(StateMain)
			assert.NoError(t, err, "run")

			refPath := replaceExt(path, ".ref")
			ref, err := os.ReadFile(refPath)
			assert.NoError(t, err, "read ref")

			if out.String() != string(ref) {
				diff := diff.CharacterDiff(out.String(), string(ref))
				t.Errorf("%q doesn't match %q\n%s", path, refPath, diff)
			}
		})

		return nil
	})
	assert.NoError(t, err, "walk testdata")
}

// replaceExt replace extension with replacement
// Used to get ref file: json.sample -> json.ref
func replaceExt(name, replacement string) string {
	ext := filepath.Ext(name)
	idx := strings.LastIndex(name, ext)
	base := name[:idx]
	return base + replacement
}
