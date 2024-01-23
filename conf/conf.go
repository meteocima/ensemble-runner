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
	GeogridProcCount          int    `yaml:"GeogridProc"`
	MetgridProcCount          int    `yaml:"MetgridProc"`
	WrfProcCount              int    `yaml:"WrfProc"`
	WrfStepProcCount          int    `yaml:"WrfStepProcCount"`
	WrfdaProcCount            int    `yaml:"WrfdaProc"`
	RealProcCount             int    `yaml:"RealProc"`
	MpiOptions                string `yaml:"MpiOptions"`
	ObDataDir                 string `yaml:"ObDataDir"`
	GeogDataDir               string `yaml:"GeogDataDir"`
	GfsDir                    string `yaml:"GfsDir"`
	CovarMatrixesDir          string `yaml:"CovarMatrixesDir"`
	RunWPS                    bool   `yaml:"RunWPS"`
	EnsembleMembers           int    `yaml:"EnsembleMembers"`
	EnsembleParallelism       int    `yaml:"EnsembleParallelism"`
	AssimilateOnlyInnerDomain bool   `yaml:"AssimilateOnlyInnerDomain"`
	AssimilateFirstCycle      bool   `yaml:"AssimilateFirstCycle"`
	CoresPerNode              int    `yaml:"CoresPerNode"`
}{}

func Initialize() {
	cfgFile := filepath.Join(folders.Rootdir, "config.yaml")
	log.Info("Reading configuration from %s", cfgFile)
	cfg := errors.CheckResult(os.ReadFile(cfgFile))
	//fmt.Printf("Configuration:\n %s\n", cfg)
	errors.Check(os.Chdir(folders.Rootdir))
	errors.Check(yaml.Unmarshal(cfg, &Values))

	for _, dir := range []*string{
		&Values.ObDataDir,
		&Values.GeogDataDir,
		&Values.GfsDir,
		&Values.CovarMatrixesDir,
	} {
		if !filepath.IsAbs(*dir) {
			*dir = errors.CheckResult(filepath.Abs(*dir))
		}
	}

	for name, value := range map[string]string{
		"GEOG_DATA": Values.GeogDataDir,
		"GFS":       Values.GfsDir,
		"BE_DIR":    Values.CovarMatrixesDir,
		"MPIOPTS":   Values.MpiOptions,
		"OB_DATDIR": Values.ObDataDir,
	} {
		errors.Check(os.Setenv(name, value))
	}

	for name, value := range map[string]any{
		"GeogridProcCount":          Values.GeogridProcCount,
		"MetgridProcCount":          Values.MetgridProcCount,
		"WrfProcCount":              Values.WrfProcCount,
		"WrfStepProcCount":          Values.WrfStepProcCount,
		"WrfdaProcCount":            Values.WrfdaProcCount,
		"RealProcCount":             Values.RealProcCount,
		"MpiOptions":                Values.MpiOptions,
		"ObDataDir":                 Values.ObDataDir,
		"GeogDataDir":               Values.GeogDataDir,
		"GfsDir":                    Values.GfsDir,
		"CovarMatrixesDir":          Values.CovarMatrixesDir,
		"RunWPS":                    Values.RunWPS,
		"EnsembleMembers":           Values.EnsembleMembers,
		"EnsembleParallelism":       Values.EnsembleParallelism,
		"AssimilateOnlyInnerDomain": Values.AssimilateOnlyInnerDomain,
		"AssimilateFirstCycle":      Values.AssimilateFirstCycle,
	} {
		log.Info("  -- %s: %v", name, value)
	}

}
