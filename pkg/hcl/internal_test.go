package hcl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

func Test_HCLFile_WriteToFile(t *testing.T) {
	type checkFn func(*testing.T, *HCLFile, string, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, f *HCLFile, outputFilePath string, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}

	fileExistsWithRightNumOfBytes := func() checkFn {
		return func(t *testing.T, f *HCLFile, outputFilePath string, err error) {
			oF, err := os.Stat(outputFilePath)
			if err != nil {
				t.Fatalf("error checking on file %s; err = %v", outputFilePath, oF)
			}

			if !oF.Mode().IsRegular() {
				t.Fatalf("%s is not a file", outputFilePath)
			}

			if int64(len(f.GetFileBytes())) != oF.Size() {
				t.Errorf("the created file (%s) does not have the same number of bytes as the HCLFile", outputFilePath)
			}
		}
	}

	tests := map[string]struct {
		numOfResources int
		checks         []checkFn
	}{
		"one resource": {
			numOfResources: 1,
			checks:         check(hasNoErr(), fileExistsWithRightNumOfBytes()),
		},
		"two resources": {
			numOfResources: 2,
			checks:         check(hasNoErr(), fileExistsWithRightNumOfBytes()),
		},
		"zero resources": {
			numOfResources: 0,
			checks:         check(hasNoErr(), fileExistsWithRightNumOfBytes()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			hf := hclwrite.NewEmptyFile()
			f := &HCLFile{
				file:     hf,
				rootBody: hf.Body(),
			}

			for i := 0; i < tc.numOfResources; i++ {
				f.file.Body().AppendNewBlock("test", []string{"t1", strconv.Itoa(i)})
			}

			outputFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.hcl", name))

			err := f.WriteToFile(outputFilePath)
			for _, check := range tc.checks {
				check(t, f, outputFilePath, err)
			}
		})
	}
}
