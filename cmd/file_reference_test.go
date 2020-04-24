package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFileReferenceCmd(t *testing.T) {
	tests := map[string]struct {
		k8sFileName string
		checks      []checkFn
		wantErr     bool
	}{
		"One file with one kubernetes object reference k8s manifest file": {
			k8sFileName: "one-obj",
			checks:      check(hasNoErr(), goldenMatchesGot()),
			wantErr:     false,
		},
		"One file with multiple kubernetes objects and inline k8s manifest": {
			k8sFileName: "multiple-objs",
			checks:      check(hasNoErr(), goldenMatchesGot()),
			wantErr:     false,
		},
		"Non existant file should return error": {
			k8sFileName: "non-existant",
			checks:      check(hasErr("no such file or directory")),
			wantErr:     true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup files
			k8sFilePath := filepath.Join(testdataFilePath, "k8s-files", fmt.Sprintf("%s.yaml", tc.k8sFileName))

			var goldenFileName string
			goldenFileName = fmt.Sprintf("%s-file-reference.tf", tc.k8sFileName)

			goldenFilePath := filepath.Join(testdataFilePath, "golden", goldenFileName)
			outputFilePath := filepath.Join(os.TempDir(), goldenFileName)

			// setup root command and persistent flags
			commandToRun := "file-reference"
			rc := rootCommandSetup(k8sFilePath, outputFilePath, commandToRun)
			// create buffer for output
			b := bytes.NewBufferString("")
			rc.SetOut(b)

			err := rc.Execute()
			if err != nil && !tc.wantErr {
				t.Errorf("running inline command failed. err = %v", err)
			}

			// if updateFlag perform the same command and write results to golden file
			if updateFlag {
				rc := rootCommandSetup(k8sFilePath, goldenFilePath, commandToRun)
				err := rc.Execute()
				if err != nil {
					t.Errorf("running inline command failed. err = %v", err)
				}
			}

			for _, check := range tc.checks {
				check(t, goldenFilePath, outputFilePath, err)
			}
		})
	}

}
