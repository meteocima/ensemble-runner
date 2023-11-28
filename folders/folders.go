package folders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var WRF string
var WPS string
var WRFDA string

var Rootdir string
var TemplatesDir string
var WorkDir string

func envVar(verbose bool, varname string) string {
	if val, ok := os.LookupEnv(varname); !ok {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ System misconfiguration: cannot read environment variable $%s\n", varname)
		}
		os.Exit(1)
		return ""
	} else {
		return val
	}
}

func PrintVer(verbose bool, varname string) {
	varvalue := os.Getenv(varname)
	cmd := exec.Command("head", "-n1", varvalue+"/README")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ Invalid %s directory: version not accessible:\n Path: %s\n Error: %s\n", varname, varvalue, string(buf))
		}
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("%s='%s'\n ✅ Found %s\n", varname, varvalue, string(buf))
	}

}

// Initialize folder vars and check environment validity
func Initialize(verbose bool) {
	WRF = envVar(verbose, "WRF_DIR")
	WPS = envVar(verbose, "WPS_DIR")
	WRFDA = envVar(verbose, "WRFDA_DIR")
	Rootdir = envVar(verbose, "WRFITA_ROOTDIR")
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

	PrintVer(verbose, "WRF_DIR")
	PrintVer(verbose, "WPS_DIR")
	PrintVer(verbose, "WRFDA_DIR")
	if verbose {
		fmt.Fprintf(os.Stderr, "WRFITA_ROOTDIR=%s\n", Rootdir)
	}

	if info, err := os.Stat(TemplatesDir); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ Invalid root directory: `templates` directory not accessible:\n Path: %s\n Error: %s\n", TemplatesDir, err)
		}
		os.Exit(1)
	} else if !info.IsDir() {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ Invalid root directory: `templates` directory exists and is not a directory.\n Path: %s", TemplatesDir)
		}
		os.Exit(1)
	}
	fmt.Printf(" ✅ Found templates directory\n")

	WorkDir = filepath.Join(Rootdir, "workdir")
	if info, err := os.Stat(WorkDir); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ Invalid root directory: `workdir` directory not accessible:\n Path: %s\n Error: %s\n", TemplatesDir, err)
		}
		os.Exit(1)
	} else if !info.IsDir() {
		if verbose {
			fmt.Fprintf(os.Stderr, "❌ Invalid root directory: `workdir` directory exists and is not a directory.\n Path: %s", WorkDir)
		}
		os.Exit(1)
	}
	if verbose {
		fmt.Printf(" ✅ Found workdir directory\n")
	}

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
