package hcl_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	h "github.com/sebradloff/rawk8stfc/pkg/hcl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	updateFlag bool
)

func init() {
	flag.BoolVar(&updateFlag, "update", false, "Set this flag to update the golden files.")
}

func TestHCLFile_K8sObjectToResourceBlock(t *testing.T) {
	type checkFn func(*testing.T, *h.HCLFile, string, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, gotFile *h.HCLFile, goldenFilePath string, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}

	resourceBlockAndLabelsCorrect := func() checkFn {
		return func(t *testing.T, gotFile *h.HCLFile, goldenFilePath string, err error) {
			numBlocksGot := len(gotFile.GetFileRootBody().Blocks())
			if numBlocksGot != 1 {
				t.Fatalf("got more than one block %d; want %d", numBlocksGot, 1)
			}

			for _, block := range gotFile.GetFileRootBody().Blocks() {
				wantType := "resource"
				if block.Type() != wantType {
					t.Errorf("block type = %s; want %s", block.Type(), wantType)
				}

				if len(block.Labels()) != 2 {
					t.Fatalf("block labels len(%d); want 2", len(block.Labels()))
				}
				wantFirstLabel := "k8s_manifest"
				if block.Labels()[0] != wantFirstLabel {
					t.Errorf("first block label = %s; want %s", block.Labels()[0], wantFirstLabel)
				}
			}
		}
	}

	goldenFileMatchesGotFile := func() checkFn {
		return func(t *testing.T, gotFile *h.HCLFile, goldenFilePath string, err error) {
			goldenFile := getGoldenFile(t, goldenFilePath)

			if !bytes.Equal(gotFile.GetFileBytes(), goldenFile.Bytes()) {
				t.Errorf("the file bytes do not match the golden file bytes (%s)", goldenFilePath)
			}

		}
	}

	contentHasSubstring := func(keyword string) checkFn {
		return func(t *testing.T, gotFile *h.HCLFile, goldenFilePath string, err error) {

			for _, block := range gotFile.GetFileRootBody().Blocks() {
				wantAttr := "content"
				attr := block.Body().GetAttribute(wantAttr)
				if attr == nil {
					t.Errorf("got no body attribute named %s", wantAttr)
				}

				contentVal := string(attr.BuildTokens(nil).Bytes())

				if !strings.Contains(contentVal, keyword) {
					t.Errorf("content value did not inclued keyword %s; got = %s", keyword, contentVal)
				}
			}
		}
	}

	tests := map[string]struct {
		object        unstructured.Unstructured
		inline        bool
		pathToK8sFile string
		checks        []checkFn
	}{
		"one object with inline content": {
			object:        testObject("one"),
			pathToK8sFile: "",
			checks:        check(hasNoErr(), resourceBlockAndLabelsCorrect(), contentHasSubstring("<<EOT"), contentHasSubstring("name: one"), goldenFileMatchesGotFile()),
		},
		"another object with inline content": {
			object:        testObject("another"),
			pathToK8sFile: "",
			checks:        check(hasNoErr(), resourceBlockAndLabelsCorrect(), contentHasSubstring("<<EOT"), contentHasSubstring("name: another"), goldenFileMatchesGotFile()),
		},
		"one object with file content": {
			object:        testObject("one"),
			pathToK8sFile: "fake-one.yaml",
			checks:        check(hasNoErr(), resourceBlockAndLabelsCorrect(), contentHasSubstring("file"), contentHasSubstring("fake-one.yaml"), goldenFileMatchesGotFile()),
		},
		"another object with file content": {
			object:        testObject("another"),
			pathToK8sFile: "fake-another.yaml",
			checks:        check(hasNoErr(), resourceBlockAndLabelsCorrect(), contentHasSubstring("file"), contentHasSubstring("fake-another.yaml"), goldenFileMatchesGotFile()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotFile := h.NewHCLFile()

			funcErr := gotFile.K8sObjectToResourceBlock(&tc.object, tc.pathToK8sFile)

			goldenFilePath := filepath.Join("test-fixtures", fmt.Sprintf("%s.hcl", strings.ReplaceAll(name, " ", "_")))
			if updateFlag {
				err := gotFile.WriteToFile(goldenFilePath)
				if err != nil {
					t.Fatalf("could not update golden file %s; err = %v", goldenFilePath, err)
				}
			}

			for _, check := range tc.checks {
				check(t, gotFile, goldenFilePath, funcErr)
			}
		})
	}
}

func testObject(name string) unstructured.Unstructured {
	var o unstructured.Unstructured
	o.SetAPIVersion("apps/v1")
	o.SetKind("Deployment")
	o.SetName(name)
	o.SetNamespace("test")
	return o
}

func getGoldenFile(t *testing.T, goldenFilePath string) *hclwrite.File {
	wantBytes, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("failed to read the goldenFile file: %s. err = %v", goldenFilePath, err)
	}

	wantFile, diags := hclwrite.ParseConfig(wantBytes, goldenFilePath, hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatalf("parsing wantFile into hclwrite.File had issues; err = %v", diags.Error())
	}

	return wantFile
}
