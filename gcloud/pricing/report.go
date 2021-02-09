package pricing

import (
	"encoding/csv"
	"os"
)

var header = []string{"resource type", "count", "spec",
	"max usage", "max cost", "projected usage", "projected cost"}

type CostReporter struct {
	file   *os.File
	writer *csv.Writer
}

func (c *CostReporter) Init(f *os.File) {
	if c.writer == nil {
		c.writer = csv.NewWriter(f)
	}
	c.file = f
	c.writer.Write(header)
}

func (c *CostReporter) AddLine(line []string) error {
	if c.writer == nil {
		c.writer = csv.NewWriter(c.file)
	}
	return c.writer.Write(line)
}

func (c *CostReporter) Flush() {
	if c.writer != nil {
		c.writer.Flush()
	}
}
