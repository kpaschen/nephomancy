package command

import (
	"testing"
)

func TestWorkingDir(t *testing.T) {
	c := Command{}
	wd, err := c.WorkingDir()
	if err != nil {
		t.Errorf("WorkingDir returned error: %v\n", err)
	}
	if wd == "" {
		t.Errorf("WorkingDir returned empty string, expected cwd\n")
	}

	c = Command{workingDirFlag: "/tmp"}
	wd, err = c.WorkingDir()
	if err != nil {
		t.Errorf("WorkingDir returned error: %v\n", err)
	}
	if wd != "/tmp" {
		t.Errorf("WorkingDir returned %s, expected /tmp\n", wd)
	}

	c = Command{workingDir: "/usr"}
	wd, err = c.WorkingDir()
	if err != nil {
		t.Errorf("WorkingDir returned error: %v\n", err)
	}
	if wd != "/usr" {
		t.Errorf("WorkingDir returned %s, expected /usr\n", wd)
	}
}

// This test will only work when the file separator is /.
func TestDataDir(t *testing.T) {
	c := Command{workingDir: "/tmp"}
	dd, err := c.DataDir()
	if err != nil {
		t.Errorf("DataDir returned error: %v\n", err)
	}
	if dd != "/tmp/.nephomancy/gcloud/data" {
		t.Errorf("DataDir returned %s, expected /tmp/.nephomancy/gcloud/data", dd)
	}
}
