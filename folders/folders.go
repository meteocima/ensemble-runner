package folders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/meteocima/ensemble-runner/log"
)

var WRF string
var WPS string
var WRFDA string

var Rootdir string
var TemplatesDir string
var WorkDir string

func envVar(varname string) string {
	if val, ok := os.LookupEnv(varname); !ok {
		log.Error("System misconfiguration: cannot read environment variable $%s\n", varname)
		os.Exit(1)
		return ""
	} else {
		return val
	}
}

func PrintVer(varname string) {
	varvalue := os.Getenv(varname)
	cmd := exec.Command("head", "-n1", varvalue+"/README")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Invalid %s directory: version not accessible:\n Path: %s\n Error: %s\n", varname, varvalue, string(buf))
		os.Exit(1)
	}
	if buf[len(buf)-1] == '\n' {
		buf = buf[0 : len(buf)-1]
	}
	log.Info("%s='%s'", varname, varvalue)
	log.Info("  -- Found %s", string(buf))

}

// Initialize folder vars and check environment validity
func Initialize() {
	WRF = envVar("WRF_DIR")
	WPS = envVar("WPS_DIR")
	WRFDA = envVar("WRFDA_DIR")
	Rootdir = envVar("WRFITA_ROOTDIR")
	Rootdir = os.ExpandEnv(Rootdir)
	TemplatesDir = filepath.Join(Rootdir, "templates")

	for _, dir := range []*string{
		&WRF,
		&WPS,
		&WRFDA,
	} {
		if !filepath.IsAbs(*dir) {
			*dir = filepath.Join(Rootdir, *dir)
		}
		*dir = os.ExpandEnv(*dir)

	}

	PrintVer("WRF_DIR")
	PrintVer("WPS_DIR")
	PrintVer("WRFDA_DIR")
	log.Info("WRFITA_ROOTDIR=%s", Rootdir)

	if info, err := os.Stat(TemplatesDir); err != nil {
		log.Error("Invalid root directory: `templates` directory not accessible:\n Path: %s\n Error: %s\n", TemplatesDir, err)
		os.Exit(1)
	} else if !info.IsDir() {
		log.Error("Invalid root directory: `templates` directory exists and is not a directory.\n Path: %s", TemplatesDir)
		os.Exit(1)
	}
	log.Info("  -- Found templates directory")

	WorkDir = filepath.Join(Rootdir, "workdir")
	if info, err := os.Stat(WorkDir); err != nil {
		log.Error("Invalid root directory: `workdir` directory not accessible:\n Path: %s\n Error: %s\n", TemplatesDir, err)
		os.Exit(1)
	} else if !info.IsDir() {
		log.Error("Invalid root directory: `workdir` directory exists and is not a directory.\n Path: %s", WorkDir)
		os.Exit(1)
	}
	log.Info("  -- Found workdir directory")

	// check for availability in path of dirprep, prepvars, chdates

}

func WPSProcWorkdir(workdir string) string {
	return filepath.Join(workdir, "wps")
}

func DAProcWorkdir(workdir string, startTime time.Time, domain int) string {
	return filepath.Join(workdir, fmt.Sprintf("da%s_d%02d", startTime.Format("15"), domain))
}

func WrfProcWorkdir(workdir string, startTime time.Time) string {
	return filepath.Join(workdir, "wrf"+startTime.Format("15"))
}
