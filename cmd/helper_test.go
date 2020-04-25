package cmd_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sebradloff/raw-k8s-tf-compiler/cmd"
	"github.com/spf13/cobra"
)

type checkFn func(*testing.T, string, string, error)

func check(fns ...checkFn) []checkFn { return fns }

func hasErr(errMsg string) checkFn {
	return func(t *testing.T, goldenFilePath, outputFilePath string, err error) {
		if err == nil {
			t.Fatal("want err; got nil")
		}

		if !strings.Contains(err.Error(), errMsg) {
			t.Errorf("error did not contain string %s; got %v", errMsg, err)
		}
	}
}

func hasNoErr() checkFn {
	return func(t *testing.T, goldenFilePath, outputFilePath string, err error) {
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}
	}
}

func goldenMatchesGot() checkFn {
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

func rootCommandSetup(k8sFilePath, outputFilePath, commandToRun string) *cobra.Command {
	rc := cmd.NewRootCmd()
	rc.PersistentFlags().Set("k8sFile", k8sFilePath)
	rc.PersistentFlags().Set("outputFile", outputFilePath)
	rc.SetArgs([]string{commandToRun})
	return rc
}
