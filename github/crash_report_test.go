package github

import (
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestStackRemoveSelfCaller(t *testing.T) {
	actual := `goroutine 1 [running]:
github.com/jingweno/gh/github.ReportCrash(0xc2000b5000, 0xc2000b49c0)
	/Users/calavera/github/go/src/github.com/jingweno/gh/github/crash_report.go:16 +0x97
github.com/jingweno/gh/commands.create(0x47f8a0, 0xc2000cf770)
	/Users/calavera/github/go/src/github.com/jingweno/gh/commands/create.go:54 +0x63
github.com/jingweno/gh/commands.(*Runner).Execute(0xc200094640, 0xc200094640, 0x21, 0xc2000b0a40)
	/Users/calavera/github/go/src/github.com/jingweno/gh/commands/runner.go:72 +0x3b7
main.main()
	/Users/calavera/github/go/src/github.com/jingweno/gh/main.go:10 +0xad`

	expected := `goroutine 1 [running]:
github.com/jingweno/gh/commands.create(0x47f8a0, 0xc2000cf770)
	/Users/calavera/github/go/src/github.com/jingweno/gh/commands/create.go:54 +0x63
github.com/jingweno/gh/commands.(*Runner).Execute(0xc200094640, 0xc200094640, 0x21, 0xc2000b0a40)
	/Users/calavera/github/go/src/github.com/jingweno/gh/commands/runner.go:72 +0x3b7
main.main()
	/Users/calavera/github/go/src/github.com/jingweno/gh/main.go:10 +0xad`

	s := formatStack([]byte(actual))
	assert.Equal(t, expected, s)
}

func TestSaveAlwaysReportOption(t *testing.T) {
	checkSavedReportCrashOption(t, true, "a", "always")
	checkSavedReportCrashOption(t, true, "always", "always")
}

func TestSaveNeverReportOption(t *testing.T) {
	checkSavedReportCrashOption(t, false, "e", "never")
	checkSavedReportCrashOption(t, false, "never", "never")
}

func TestDoesntSaveYesReportOption(t *testing.T) {
	checkSavedReportCrashOption(t, false, "y", "")
	checkSavedReportCrashOption(t, false, "yes", "")
}

func TestDoesntSaveNoReportOption(t *testing.T) {
	checkSavedReportCrashOption(t, false, "n", "")
	checkSavedReportCrashOption(t, false, "no", "")
}

func checkSavedReportCrashOption(t *testing.T, always bool, confirm, expected string) {
	file, _ := ioutil.TempFile("", "test-gh-config-")
	defer os.RemoveAll(file.Name())
	defaultConfigsFile = file.Name()

	ccreds := Credentials{Host: "github.com", User: "jingweno", AccessToken: "123"}
	c := Configs{Credentials: []Credentials{ccreds}}

	saveReportConfiguration(&c, confirm, always)

	cc := &Configs{}
	err := loadFrom(file.Name(), cc)
	assert.Equal(t, nil, err)

	assert.Equal(t, expected, cc.ReportCrash)
}