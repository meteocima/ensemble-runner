package conf

import (
	"os"
	"path/filepath"

	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
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
		log.Info("Reading configuration from %s", cfgFile)
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

	for name, value := range map[string]any{
		"GeogridProcCount": Values.GeogridProcCount,
		"MetgridProcCount": Values.MetgridProcCount,
		"WrfProcCount":     Values.WrfProcCount,
		"WrfdaProcCount":   Values.WrfdaProcCount,
		"RealProcCount":    Values.RealProcCount,
		"MpiOptions":       Values.MpiOptions,
		"ObDataDir":        Values.ObDataDir,
		"GeogDataDir":      Values.GeogDataDir,
		"GfsDir":           Values.GfsDir,
		"BeDir":            Values.BeDir,
		"TemplatesDir":     Values.TemplatesDir,
		"Workdir":          Values.Workdir,
		"Bindir":           Values.Bindir,
		"PostprocRules":    Values.PostprocRules,
	} {
		log.Info("  -- %s: %v", name, value)
	}

}
