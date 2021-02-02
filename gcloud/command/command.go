package command

import (
	"database/sql"
	"flag"
	"os"
	"path/filepath"
)

type Command struct {
	// All relative paths are relative to this directory.
	// Defaults to current working directory but can be overridden
	// with the --workingdir flag.
	workingDir string

	// The directory where the database file and other internals are kept.
	dataDir string

	// The filename where the database is stored.
	dbFile string

	// The gcloud project to pass into APIs that require a project.
	Project string

	// This is the original working dir flag value. I should probably only initialize
	// the workingDir value once I've verified the directory exists.
	workingDirFlag string

	// Internal handle for sqlite database holding the sku cache.
	dbHandle *sql.DB

	// File to load an asset model from.
	modelInFile string

	// File to save an asset model to.
	modelOutFile string

	// File to write cost report to.
	costReportOutFile string
}

// Create a flag set with flags common to all commands.
func (c *Command) defaultFlagSet(cn string) *flag.FlagSet {
	f := flag.NewFlagSet(cn, flag.ExitOnError)
	f.StringVar(&c.workingDirFlag, "workingdir", "", "Working Directory. Defaults to current working directory.")
	f.StringVar(&c.Project, "project", "", "The name of the gcloud project to use for API access.")
	f.StringVar(&c.modelOutFile, "modelout", "", "Where to write the asset model to (json protobuf).")
	f.StringVar(&c.modelInFile, "modelin", "", "Where to read the asset model from (json protobuf).")
	return f
}

func (c *Command) WorkingDir() (string, error) {
	if c.workingDir != "" {
		return c.workingDir, nil
	}
	if c.workingDirFlag != "" {
		c.workingDir = c.workingDirFlag
		return c.workingDir, nil
	}
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	c.workingDir = pwd
	return c.workingDir, nil
}

func (c *Command) DataDir() (string, error) {
	if c.dataDir != "" {
		return c.dataDir, nil
	}
	wd, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	dd := filepath.Join(wd, ".nephomancy", "gcloud", "data")
	err = os.MkdirAll(dd, 0777)
	if err != nil {
		return "", err
	}
	c.dataDir = dd
	return c.dataDir, nil
}

func (c *Command) DbFile() (string, error) {
	dd, err := c.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dd, "sku-cache.db"), nil
}

func (c *Command) ModelInFile() (string, error) {
	if c.modelInFile == "" {
		return "", nil
	}
	wd, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, c.modelInFile), nil
}

func (c *Command) ModelOutFile() (string, error) {
	if c.modelOutFile == "" {
		return "", nil
	}
	wd, err  := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, c.modelOutFile), nil
}

func (c *Command) DbHandle() (*sql.DB, error) {
	if c.dbHandle != nil {
		return c.dbHandle, nil
	}
	f, err := c.DbFile()
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (c *Command) CloseDb() error {
	if c.dbHandle == nil {
		return nil
	}
	err := c.dbHandle.Close()
	if err != nil {
		return err
	}
	return nil
}
