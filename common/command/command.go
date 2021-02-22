package command

import (
	"flag"
	"fmt"
	"github.com/kennygrant/sanitize"
	"google.golang.org/protobuf/encoding/protojson"
	"io/ioutil"
	"log"
	"nephomancy/common/resources"
	"os"
	"path/filepath"
	"strings"
)

const workingDirDoc = `Path to working directory. Defaults to current working directory. A data directory called .nephomancy/data will be created underneath this working directory if it does not exist yet.`

const projectInDoc = `Filename to read project information from. If this flag is not set, a generic template project will be created. If this flag is set, the project is expected to be represented as a json-encoded protocol buffer.`

const projectOutDoc = `Filename to save project information to. Project information will be written as a json-encoded protocol buffer.`

const costReportDoc = `Filename to save cost report (csv) to.`

const providerDoc = `Name of a cloud provider. A registry entry must exist for this provider. Supported providers are: gcloud, green.ch, custom.`

type Command struct {
	// All relative paths are relative to this directory.
	// Defaults to current working directory but can be overridden
	// with the --workingdir flag.
	workingDir string

	// This is the original working dir flag value. I should probably only initialize
	// the workingDir value once I've verified the directory exists.
	workingDirFlag string

	// Directory with the cached price data.
	dataDir string

	// File to load a project from.
	projectInFile string

	// File to save the project to.
	projectOutFile string

	// File to write cost report to.
	costReportFile string

	// The provider to use.
	provider string
}

// Create a flag set with flags common to most commands.
func (c *Command) DefaultFlagSet(cn string) *flag.FlagSet {
	f := flag.NewFlagSet(cn, flag.ExitOnError)
	f.StringVar(&c.workingDirFlag, "workingdir", "", "Working Directory. Defaults to current working directory.")
	f.StringVar(&c.projectOutFile, "projectout", "", "Where to save the project (json protobuf).")
	f.StringVar(&c.projectInFile, "projectin", "", "Where to read the project from (json protobuf).")
	f.StringVar(&c.costReportFile, "costreport", "", "Where to write the cost report (csv).")
	f.StringVar(&c.provider, "provider", "", "Provider")
	return f
}

func (c *Command) WorkingDir() (string, error) {
	if c.workingDir != "" {
		return c.workingDir, nil
	}
	if c.workingDirFlag != "" {
		c.workingDir = filepath.Clean(c.workingDirFlag)
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
	dd := filepath.Join(wd, ".nephomancy", "data")
	err = os.MkdirAll(dd, 0777)
	if err != nil {
		return "", err
	}
	c.dataDir = dd
	return c.dataDir, nil
}

func (c *Command) ProjectInFile() (string, error) {
	if c.projectInFile == "" {
		return "", nil
	}
	wd, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	infile := sanitize.Name(c.projectInFile)
	return filepath.Join(wd, infile), nil
}

func (c *Command) ProjectOutFile(fallback string) (string, error) {
	fname := c.projectOutFile
	if fname == "" {
		fname = fallback
	}
	outfile := sanitize.Name(fname)
	wd, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, outfile), nil
}

func (c *Command) CostReportFile(fallback string) (string, error) {
	fname := c.costReportFile
	if fname == "" {
		fname = fallback
	}
	outfile := sanitize.Name(fname)
	parts := strings.Split(outfile, ".")
	if len(parts) <= 1 || parts[len(parts)-1] != "csv" {
		outfile += ".csv"
	}
	wd, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, outfile), nil
}

func (c *Command) loadProject() (*resources.Project, error) {
	infile, err := c.ProjectInFile()
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
	p := &resources.Project{}
	if err = protojson.Unmarshal(pdata, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Command) saveProject(p *resources.Project) error {
	if p.Name == "" {
		return fmt.Errorf("project %v has no name", *p)
	}
	fallback := sanitize.Name(fmt.Sprintf("%s.json", p.Name))
	outfile, err := c.ProjectOutFile(fallback)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(outfile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	options := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}
	f.WriteString(options.Format(p))
	log.Printf("Project %s saved to file %s\n", p.Name, outfile)
	return nil
}

func (c *Command) getCostFile(fallback string) (*os.File, error) {
	fname, err := c.CostReportFile(fallback)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return file, nil
}
