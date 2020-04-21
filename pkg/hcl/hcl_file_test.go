package hcl_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
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
	// given
	f := h.NewHCLFile()
	var o unstructured.Unstructured
	o.SetAPIVersion("apps/v1")
	o.SetKind("Deployment")
	o.SetName("test-1")
	o.SetNamespace("test-1")
	//when
	err := f.K8sObjectToResourceBlock(&o)
	if err != nil {
		t.Fatalf("asdfsdf %v", err)
	}

	//then
	goldenFile := filepath.Join("test-fixtures", fmt.Sprintf("%s.hcl", "test-1"))
	if updateFlag {
		err := f.WriteToFile(goldenFile)
		if err != nil {
			t.Fatalf("could not update golden file %s; err = %v", goldenFile, err)
		}
	}

	wantBytes, err := ioutil.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("failed to read the goldenFile file: %s. err = %v", goldenFile, err)
	}

	wF, diags := hclwrite.ParseConfig(wantBytes, goldenFile, hcl.InitialPos)
	if diags.HasErrors() {
		for _, diag := range diags {
			if diag.Subject != nil {
				fmt.Printf("[%s:%d] %s: %s", diag.Subject.Filename, diag.Subject.Start.Line, diag.Summary, diag.Detail)
			} else {
				fmt.Printf("%s: %s", diag.Summary, diag.Detail)
			}
		}
	}

	f.GetFileRootBody()

	if len(f.GetFileRootBody().Blocks()) != 1 {
		t.Fatalf("got more than one block %d; want %d", len(f.GetFileRootBody().Blocks()), 1)
	}

	for _, block := range f.GetFileRootBody().Blocks() {
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

		wantAttr := "content"
		attr := block.Body().GetAttribute(wantAttr)
		if attr == nil {
			t.Errorf("got no body attribue; want attribute %s", wantAttr)
		}
	}

	if !bytes.Equal(f.GetFileBytes(), wF.Bytes()) {
		t.Errorf("the file bytes do not match the golden file bytes (%s)", goldenFile)
	}
}
