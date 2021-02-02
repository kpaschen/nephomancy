package command

import (
	"database/sql"
	"flag"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	common "nephomancy/common/resources"
	"nephomancy/gcloud/assets"
	"github.com/kennygrant/sanitize"
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
		c.workingDir = sanitize.Name(c.workingDirFlag)
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
	infile := sanitize.Name(c.modelInFile)
	return filepath.Join(wd, infile), nil
}

func (c *Command) ModelOutFile(fallback string) (string, error) {
	fname := c.modelOutFile
	if fname == "" {
		fname = fallback
	}
	outfile := sanitize.Name(fname)
	wd, err  := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, outfile), nil
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

func (c *Command) loadProject() (*common.Project, error) {
	infile, err := c.ModelInFile()
	if err != nil {
		return nil, err
	}
	if infile == "" {
		return nil, nil
	}
	pdata, err := ioutil.ReadFile(infile)
	if err != nil {
		return nil, err
	}
	p, err := assets.UnmarshalProject(pdata)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Command) saveProject(p *common.Project) error {
	fallback := sanitize.Name(fmt.Sprintf("%s.json", p.Name))
	outfile, err := c.ModelOutFile(fallback)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(outfile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent: "  ",
	}
	f.WriteString(options.Format(p))
	log.Printf("Project %s saved to file %s\n", p.Name, outfile)
	return nil
}
