package cmd_test

import (
	"flag"
	"reflect"
	"sort"
	"testing"

	"github.com/sebradloff/rawk8stfc/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	updateFlag bool
)

func init() {
	flag.BoolVar(&updateFlag, "update", false, "Set this flag to update the golden files.")
}

const testdataFilePath = "../testdata"

func TestRootCmd(t *testing.T) {
	rc := cmd.NewRootCmd()

	// requires flags
	requiredFlags := &[]string{"outputFile", "k8sFile"}

	rc.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if contains(*requiredFlags, f.Name) {
			v, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
			if !ok {
				t.Errorf("required annotation not found for flag %s", f.Name)
			}
			if reflect.DeepEqual(v, []bool{true}) {
				t.Errorf("flag %s is not marked as required; got = %s", f.Name, v)
			}

			requiredFlags = remove(*requiredFlags, f.Name)
		}
	})

	if len(*requiredFlags) != 0 {
		t.Errorf("the following persistent flags were not marked as required: %s", *requiredFlags)
	}

	// has sub commands
	expectedSubCommands := &[]string{"inline", "file-reference"}
	for _, c := range rc.Commands() {
		if contains(*expectedSubCommands, c.Name()) {
			expectedSubCommands = remove(*expectedSubCommands, c.Name())
		}
	}

	if len(*expectedSubCommands) != 0 {
		t.Errorf("the following sub commands were not present on the root cmd: %s", *expectedSubCommands)
	}

}

func contains(s []string, searchterm string) bool {
	sort.Strings(s)
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func remove(s []string, searchterm string) *[]string {
	sort.Strings(s)
	i := sort.SearchStrings(s, searchterm)
	s[i] = s[len(s)-1]
	s[len(s)-1] = ""
	s = s[:len(s)-1]
	return &s
}
