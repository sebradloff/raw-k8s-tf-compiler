package cmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebradloff/rawk8stfc/cmd"
)

func TestFileReferenceCmd(t *testing.T) {
	type checkFn func(*testing.T, string, string, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, goldenFilePath, outputFilePath string, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}

	goldenMatchesGot := func() checkFn {
		return func(t *testing.T, goldenFilePath, outputFilePath string, err error) {
			got, err := os.Open(outputFilePath)
			if err != nil {
				t.Fatalf("failed to open the got file: %s. err = %v", outputFilePath, err)
			}
			defer got.Close()
			gotBytes, err := ioutil.ReadAll(got)
			if err != nil {
				t.Fatalf("failed to read the got file: %s. err = %v", outputFilePath, err)
			}

			want, err := os.Open(goldenFilePath)
			if err != nil {
				t.Fatalf("failed to open the golden file: %s. err = %v", goldenFilePath, err)
			}
			defer want.Close()
			wantBytes, err := ioutil.ReadAll(want)
			if err != nil {
				t.Fatalf("failed to read the golden file: %s. err = %v", goldenFilePath, err)
			}

			if bytes.Compare(gotBytes, wantBytes) != 0 {
				t.Fatalf("golden file (%s) does not match the output file (%s)", goldenFilePath, outputFilePath)
			}
			os.Remove(outputFilePath)
		}
	}

	tests := map[string]struct {
		k8sFileName string
		checks      []checkFn
	}{
		"One file with one kubernetes object reference k8s manifest file": {
			k8sFileName: "one-obj",
			checks:      check(hasNoErr(), goldenMatchesGot()),
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
			rc := cmd.NewRootCmd()
			rc.PersistentFlags().Set("k8sFile", k8sFilePath)
			rc.PersistentFlags().Set("outputFile", outputFilePath)
			// sub command to call
			rc.SetArgs([]string{"file-reference"})
			err := rc.Execute()
			if err != nil {
				t.Errorf("running inline command failed. err = %v", err)
			}

			// if updateFlag perform the same command and write results to golden file
			if updateFlag {
				rc.PersistentFlags().Set("k8sFile", k8sFilePath)
				rc.PersistentFlags().Set("outputFile", outputFilePath)
				// sub command to call
				rc.SetArgs([]string{"file-reference"})
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
