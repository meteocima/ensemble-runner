package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"gopkg.in/yaml.v3"
)

var Values = struct {
	GeogridProcCount int               `yaml:"GeogridProc"`
	MetgridProcCount int               `yaml:"MetgridProc"`
	WrfProcCount     int               `yaml:"WrfProc"`
	WrfdaProcCount   int               `yaml:"WrfdaProc"`
	RealProcCount    int               `yaml:"RealProc"`
	MpiOptions       string            `yaml:"MpiOptions"`
	ObDataDir        string            `yaml:"ObDataDir"`
	GeogDataDir      string            `yaml:"GeogDataDir"`
	GfsDir           string            `yaml:"GfsDir"`
	BeDir            string            `yaml:"BeDir"`
	TemplatesDir     string            `yaml:"TemplatesDir"`
	Workdir          string            `yaml:"Workdir"`
	Bindir           string            `yaml:"Bindir"`
	PostprocRules    map[string]string `yaml:"PostprocRules"`
}{}

func Initialize(verbose bool) {
	cfgFile := filepath.Join(folders.Rootdir, "config.yaml")
	if verbose {
		fmt.Printf("Reading configuration from %s\n", cfgFile)
	}
	cfg := errors.CheckResult(os.ReadFile(cfgFile))
	//fmt.Printf("Configuration:\n %s\n", cfg)
	errors.Check(os.Chdir(folders.Rootdir))
	errors.Check(yaml.Unmarshal(cfg, &Values))

	for _, dir := range []*string{
		&Values.ObDataDir,
		&Values.GeogDataDir,
		&Values.GfsDir,
		&Values.BeDir,
		&Values.TemplatesDir,
		&Values.Workdir,
		&Values.Bindir,
	} {
		if !filepath.IsAbs(*dir) {
			*dir = errors.CheckResult(filepath.Abs(*dir))
		}
	}

	for name, value := range map[string]string{
		"GEOG_DATA": Values.GeogDataDir,
		"GFS":       Values.GfsDir,
		"BE_DIR":    Values.BeDir,
		"MPIOPTS":   Values.MpiOptions,
		"OB_DATDIR": Values.ObDataDir,
	} {
		errors.Check(os.Setenv(name, value))
	}
	if verbose {
		fmt.Printf(`
	GeogridProcCount : %d
	MetgridProcCount : %d
	WrfProcCount     : %d
	WrfdaProcCount   : %d
	RealProcCount    : %d
	MpiOptions       : %s
	ObDataDir        : %s
	GeogDataDir      : %s
	GfsDir           : %s
	BeDir            : %s
	TemplatesDir     : %s
	Workdir          : %s
	Bindir           : %s
`,
			Values.GeogridProcCount,
			Values.MetgridProcCount,
			Values.WrfProcCount,
			Values.WrfdaProcCount,
			Values.RealProcCount,
			Values.MpiOptions,
			Values.ObDataDir,
			Values.GeogDataDir,
			Values.GfsDir,
			Values.BeDir,
			Values.TemplatesDir,
			Values.Workdir,
			Values.Bindir,
		)
	}
}
