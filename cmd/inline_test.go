package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInlineCmd(t *testing.T) {
	tests := map[string]struct {
		k8sFileName string
		checks      []checkFn
	}{
		"One file with one kubernetes object inline k8s manifest": {
			k8sFileName: "one-obj",
			checks:      check(hasNoErr(), goldenMatchesGot()),
		},
		"One file with multiple kubernetes objects and inline k8s manifest": {
			k8sFileName: "multiple-objs",
			checks:      check(hasNoErr(), goldenMatchesGot()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup files
			k8sFilePath := filepath.Join(testdataFilePath, "k8s-files", fmt.Sprintf("%s.yaml", tc.k8sFileName))

			var goldenFileName string
			goldenFileName = fmt.Sprintf("%s-inline.tf", tc.k8sFileName)

			goldenFilePath := filepath.Join(testdataFilePath, "golden", goldenFileName)
			testOutputFilePath := filepath.Join(os.TempDir(), goldenFileName)

			// setup root command and persistent flags
			commandToRun := "inline"
			rc := rootCommandSetup(k8sFilePath, testOutputFilePath, commandToRun)
			err := rc.Execute()
			if err != nil {
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
				check(t, goldenFilePath, testOutputFilePath, err)
			}
		})
	}

}
