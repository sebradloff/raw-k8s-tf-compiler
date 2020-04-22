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

func TestRoot(t *testing.T) {
	tests := map[string]struct {
		k8sFileName string
		inlineK8s   bool
	}{
		"One file with one kubernetes object inline k8s manifest": {
			k8sFileName: "one-obj",
			inlineK8s:   true,
		},
		"One file with multiple kubernetes objects and inline k8s manifest": {
			k8sFileName: "multiple-objs",
			inlineK8s:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup files
			k8sFilePath := filepath.Join("../testdata", "k8s-files", fmt.Sprintf("%s.yaml", tc.k8sFileName))

			var goldenFileName string
			if tc.inlineK8s {
				goldenFileName = fmt.Sprintf("%s-inline.tf", tc.k8sFileName)
			} else {
				goldenFileName = fmt.Sprintf("%s-file.tf", tc.k8sFileName)
			}

			goldenFilePath := filepath.Join("../testdata", "golden", goldenFileName)
			outputFilePath := filepath.Join(os.TempDir(), goldenFileName)

			// set cmd flags
			cmd := cmd.TestCmd()
			cmd.Flags().Set("k8sFile", k8sFilePath)
			cmd.Flags().Set("outputFile", outputFilePath)
			cmd.Flags().Set("inlineK8s", strconv.FormatBool(tc.inlineK8s))
			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("running root command failed. err = %v", err)
			}

			// if updateFlag perform the same command and write results to golden file
			if updateFlag {
				cmd.Flags().Set("k8sFile", k8sFilePath)
				cmd.Flags().Set("outputFile", goldenFilePath)
				cmd.Flags().Set("inlineK8s", strconv.FormatBool(tc.inlineK8s))
				err := cmd.RunE(cmd, []string{})
				if err != nil {
					t.Errorf("running root command failed. err = %v", err)
				}
			}

			filesAreSame(t, goldenFilePath, outputFilePath)
		})
	}

}

func filesAreSame(t *testing.T, goldenFilePath, outputFilePath string) {
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
