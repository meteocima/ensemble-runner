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
	// GeogridProcCount is the number of cores to use for geogrid.exe
	GeogridProcCount int `yaml:"GeogridProc"`
	// MetgridProcCount is the number of cores to use for metgrid.exe
	MetgridProcCount int `yaml:"MetgridProc"`
	// WrfProcCount is the number of cores to use for wrf.exe in the
	// control forecast and for ensemble members.
	WrfProcCount int `yaml:"WrfProc"`
	// WrfStepProcCount is the number of cores to use for wrf.exe in the
	// assimilation cycles.
	WrfStepProcCount int `yaml:"WrfStepProcCount"`
	// WrfdaProcCount is the number of cores to use for da_wrfvar.exe process.
	WrfdaProcCount int `yaml:"WrfdaProc"`
	// RealProcCount is the number of cores to use for real.exe
	RealProcCount int `yaml:"RealProc"`
	// MpiOptions contains additional options to pass to the mpirun command
	// when running the WRF executables.
	MpiOptions string `yaml:"MpiOptions"`
	// ObDataDir is the directory where the observation data is stored.
	ObDataDir string `yaml:"ObDataDir"`
	// GeogDataDir is the directory where the input geogrid static data is stored.
	GeogDataDir string `yaml:"GeogDataDir"`
	// GfsDir is the directory where the input GFS data is stored.
	GfsDir string `yaml:"GfsDir"`
	// CovarMatrixesDir is the directory where the background error covariance data are stored
	CovarMatrixesDir string `yaml:"CovarMatrixesDir"`
	// Whever to run preprocessing step. If false, the WPS output files are expected to be already present
	// inside 'inputs' directory. Otherwise, the WPS executables are run to generate the input files, using
	// the data in 'GfsDir' and 'GeogDataDir' as inputs.
	RunWPS bool `yaml:"RunWPS"`
	// EnsembleMembers is the number of ensemble members to run. If 0, only the control
	// forecast is run. This number does not include the control forecast.
	EnsembleMembers int `yaml:"EnsembleMembers"`

	// EnsembleParallelism contains the number of ensemble members to run in parallel
	// The main control forecast is scheduled taking into accounts this value for parallelism,
	// so `EnsembleParallelism` must be at least 1, even when no ensemble members is needed.
	EnsembleParallelism int `yaml:"EnsembleParallelism"`

	// Whether to assimilate observations or not.
	AssimilateObservations bool `yaml:"AssimilateObservations"`
	// Whether to assimilate observations only in the inner domain, or in the outer ones too.
	AssimilateOnlyInnerDomain bool `yaml:"AssimilateOnlyInnerDomain"`
	// Whether to assimilate observations only in the first cycle, or in each one of them.
	AssimilateFirstCycle bool `yaml:"AssimilateFirstCycle"`
	// Number of cores per node in the cluster where the simulation is run.
	// This is used to calculate which nodes to use for each one of the ensemble members.
	CoresPerNode int `yaml:"CoresPerNode"`
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
