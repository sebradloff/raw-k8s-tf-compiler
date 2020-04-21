package cmd_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	}{
		"One file with one kubernetes object": {
			k8sFileName: "one-obj",
		},
		"One file with multiple kubernetes objects": {
			k8sFileName: "multiple-objs",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			k8sFile := filepath.Join("../testdata", "k8s-files", fmt.Sprintf("%s.yaml", tc.k8sFileName))
			outputFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.tf", tc.k8sFileName))
			goldenFile := filepath.Join("../testdata", "golden", fmt.Sprintf("%s.tf", tc.k8sFileName))

			cmd := cmd.TestCmd()
			cmd.Flags().Set("k8sFile", k8sFile)
			cmd.Flags().Set("outputFile", outputFile)
			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Errorf("running root command failed. err = %v", err)
			}

			if updateFlag {
				cmd.Flags().Set("k8sFile", k8sFile)
				cmd.Flags().Set("outputFile", goldenFile)
				err := cmd.RunE(cmd, []string{})
				if err != nil {
					t.Errorf("running root command failed. err = %v", err)
				}
			}

			got, err := os.Open(outputFile)
			if err != nil {
				t.Fatalf("failed to open the got file: %s. err = %v", outputFile, err)
			}
			defer got.Close()
			gotBytes, err := ioutil.ReadAll(got)
			if err != nil {
				t.Fatalf("failed to read the got file: %s. err = %v", outputFile, err)
			}

			want, err := os.Open(goldenFile)
			if err != nil {
				t.Fatalf("failed to open the golden file: %s. err = %v", goldenFile, err)
			}
			defer want.Close()
			wantBytes, err := ioutil.ReadAll(want)
			if err != nil {
				t.Fatalf("failed to read the golden file: %s. err = %v", goldenFile, err)
			}

			if bytes.Compare(gotBytes, wantBytes) != 0 {
				t.Fatalf("golden file (%s) does not match the output file (%s)", goldenFile, outputFile)
			}
			os.Remove(outputFile)
		})
	}

}
