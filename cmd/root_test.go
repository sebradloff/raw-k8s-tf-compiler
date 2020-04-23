package cmd_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sebradloff/rawk8stfc/cmd"
)

var (
	updateFlag bool
)

func init() {
	flag.BoolVar(&updateFlag, "update", false, "Set this flag to update the golden files.")
}

const testdataFilePath = "../testdata"

func TestRoot(t *testing.T) {
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
		k8sFileName   string
		contentInline bool
		checks        []checkFn
	}{
		"One file with one kubernetes object inline k8s manifest": {
			k8sFileName:   "one-obj",
			contentInline: true,
			checks:        check(hasNoErr(), goldenMatchesGot()),
		},
		"One file with multiple kubernetes objects and inline k8s manifest": {
			k8sFileName:   "multiple-objs",
			contentInline: true,
			checks:        check(hasNoErr(), goldenMatchesGot()),
		},
		"One file with one kubernetes object reference k8s manifest file": {
			k8sFileName:   "one-obj",
			contentInline: false,
			checks:        check(hasNoErr(), goldenMatchesGot()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup files
			k8sFilePath := filepath.Join(testdataFilePath, "k8s-files", fmt.Sprintf("%s.yaml", tc.k8sFileName))

			var goldenFileName string
			if tc.contentInline {
				goldenFileName = fmt.Sprintf("%s-inline.tf", tc.k8sFileName)
			} else {
				goldenFileName = fmt.Sprintf("%s-file.tf", tc.k8sFileName)
			}

			goldenFilePath := filepath.Join(testdataFilePath, "golden", goldenFileName)
			outputFilePath := filepath.Join(os.TempDir(), goldenFileName)

			// set cmd flags
			cmd := cmd.TestCmd()
			cmd.Flags().Set("k8sFile", k8sFilePath)
			cmd.Flags().Set("outputFile", outputFilePath)
			cmd.Flags().Set("contentInline", strconv.FormatBool(tc.contentInline))
			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("running root command failed. err = %v", err)
			}

			// if updateFlag perform the same command and write results to golden file
			if updateFlag {
				cmd.Flags().Set("k8sFile", k8sFilePath)
				cmd.Flags().Set("outputFile", goldenFilePath)
				cmd.Flags().Set("contentInline", strconv.FormatBool(tc.contentInline))
				err := cmd.RunE(cmd, []string{})
				if err != nil {
					t.Errorf("running root command failed. err = %v", err)
				}
			}

			for _, check := range tc.checks {
				check(t, goldenFilePath, outputFilePath, err)
			}
		})
	}

}
