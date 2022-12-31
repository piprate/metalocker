// Copyright © 2016 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// The code below is borrowed from github.com/spf13/viper

package utils_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/piprate/metalocker/utils"
)

func TestAbsPathify(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip test on Windows")
	}

	home := UserHomeDir()
	homer := filepath.Join(home, "homer")
	wd, _ := os.Getwd()

	t.Setenv("HOMER_ABSOLUTE_PATH", homer)
	t.Setenv("VAR_WITH_RELATIVE_PATH", "relative")

	tests := []struct {
		input  string
		output string
	}{
		{"", wd},
		{"sub", filepath.Join(wd, "sub")},
		{"./", wd},
		{"./sub", filepath.Join(wd, "sub")},
		{"$HOME", home},
		{"$HOME/", home},
		{"$HOME/sub", filepath.Join(home, "sub")},
		{"$HOMER_ABSOLUTE_PATH", homer},
		{"$HOMER_ABSOLUTE_PATH/", homer},
		{"$HOMER_ABSOLUTE_PATH/sub", filepath.Join(homer, "sub")},
		{"$VAR_WITH_RELATIVE_PATH", filepath.Join(wd, "relative")},
		{"$VAR_WITH_RELATIVE_PATH/", filepath.Join(wd, "relative")},
		{"$VAR_WITH_RELATIVE_PATH/sub", filepath.Join(wd, "relative", "sub")},
	}

	for _, test := range tests {
		got := AbsPathify(test.input)
		if got != test.output {
			t.Errorf("Got %v\nexpected\n%q", got, test.output)
		}
	}
}
