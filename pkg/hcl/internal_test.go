package hcl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func Test_generateResourceName(t *testing.T) {
	tests := map[string]struct {
		name    string
		obj     unstructured.Unstructured
		want    string
		wantErr bool
	}{
		"no special characters": {
			obj:     testObject("test", "v1", "Deployment", "test-ns"),
			want:    "test-ns_v1_Deployment_test",
			wantErr: false,
		},
		"non valid k8s character now and not valid hcl": {
			obj:     testObject("test", "v1", "Deployment", "test@ns"),
			want:    "test@ns_v1_Deployment_test",
			wantErr: true,
		},
		"valid k8s backslash character replace with dash": {
			obj:     testObject("test", "apps/v1", "Deployment", "test-ns"),
			want:    "test-ns_apps-v1_Deployment_test",
			wantErr: false,
		},
		"valid k8s namespace of only numbers": {
			obj:     testObject("test", "apps/v1", "Deployment", "123"),
			want:    "n_123_apps-v1_Deployment_test",
			wantErr: false,
		},
		"if no k8s namespace insert default": {
			obj:     testObject("test", "v1", "Deployment", ""),
			want:    "default_v1_Deployment_test",
			wantErr: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := generateResourceName(&tc.obj)
			if (err != nil) && !tc.wantErr {
				t.Fatalf("generateResourceName() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("generateResourceName() = %v, want %v", got, tc.want)
			}
		})
	}
}

func testObject(name, apiVersion, kind, namespace string) unstructured.Unstructured {
	var o unstructured.Unstructured
	o.SetAPIVersion(apiVersion)
	o.SetKind(kind)
	o.SetName(name)
	o.SetNamespace(namespace)
	return o
}
