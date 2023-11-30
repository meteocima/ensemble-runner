package server

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	pt "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/meteocima/ensemble-runner/errors"
	"github.com/meteocima/ensemble-runner/folders"
	"github.com/meteocima/ensemble-runner/log"
)

func MkdirAll(dir string, mod fs.FileMode) {
	errors.Check(os.MkdirAll(dir, 0775))
}

func Rmdir(dir string) {
	errors.Check(os.RemoveAll(dir))
}

func CopyFile(workdir, src, dst string) {
	srcRel := errors.CheckResult(filepath.Rel(workdir, src))
	dstRel := errors.CheckResult(filepath.Rel(workdir, dst))
	log.Debug(" - Copying file $WORKDIR/%s to $WORKDIR/%s", srcRel, dstRel)
	bytesRead := errors.CheckResult(os.ReadFile(src))
	errors.Check(os.WriteFile(dst, bytesRead, 0664))
}

func ExecRetry(cmd, cwd, collectStdErr, logsToSave string) {
	var err error
	var g glob.Glob
	if logsToSave != "" {
		g = glob.MustCompile(logsToSave)
	}

	for i := 0; i < 5; i++ {
		err = tryExec(cmd, cwd, collectStdErr)
		if err == nil {
			break
		}
		log.Warning("Command `%s` has failed: %s. Retry n.%d in 1 minute", cmd, err.Error(), i+1)

		if logsToSave != "" {
			files, err := os.ReadDir(cwd)
			if err != nil {
				log.Warning("Cannot save logs for previous attempt: %s", err.Error())
			} else {
				for _, f := range files {
					if !g.Match(f.Name()) {
						continue
					}
					input, err := os.ReadFile(pt.Join(cwd, f.Name()))
					if err != nil {
						log.Warning("Cannot read original log file %s: %s", f.Name(), err.Error())
						continue
					}

					destinationFile := fmt.Sprintf("%s.%d", f.Name(), i)
					err = os.WriteFile(pt.Join(cwd, destinationFile), input, 0644)
					if err != nil {
						log.Warning("Cannot save log file %s to %s: %s", f.Name(), destinationFile, err.Error())
					}

				}
			}
		}

		time.Sleep(1 * time.Minute)
	}
	errors.Check(err)
}

func Exec(cmd, cwd, collectStdErr string) {
	errors.Check(tryExec(cmd, cwd, collectStdErr))
}

func tryExec(cmd, cwd, collectStdErr string) error {
	var log *os.File
	if collectStdErr != "" {

		l, err := os.Create(filepath.Join(cwd, collectStdErr))
		if err != nil {
			return fmt.Errorf("cannot write log file %s: %s", collectStdErr, err)
		}
		defer l.Close()
		log = l
	}

	if !filepath.IsAbs(cwd) {
		cwd = errors.CheckResult(filepath.Abs(cwd))
	}

	c := exec.Command("bash", "-c", cmd)
	c.Dir = cwd
	c.Stdout = log
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return err
	}

	var stderrChuncks []string
	go func() {
		var buf [1024]byte
		var n int
		for {
			n, err = stderrPipe.Read(buf[:])

			if err == io.EOF {
				err = nil
			} else if err != nil {
				log.Write([]byte(fmt.Sprintf("\nERROR: cannot read stderr: %s\n", err.Error())))
				return
			}

			if n == 0 {
				return
			}

			log.Write(buf[0:n])

			stderrChuncks = append(stderrChuncks, string(buf[0:n]))
		}
	}()

	if err = c.Run(); err != nil {
		return fmt.Errorf(
			"command failed: cannot start:\n"+
				"    => cmd: %s\n"+
				"    => wdir: %s\n"+
				"    => err: %w\n"+
				"    => stderr: %s\n"+
				"    ==",
			cmd, cwd, err, strings.Join(stderrChuncks, ""),
		)
	}
	return nil
}

func RenderTemplate(targetDir, name string, startDate time.Time, durationHours int) {
	defer errors.OnFailuresWrap("cannot render template directory `%s` to `%s`: %w", name, targetDir)
	Exec(fmt.Sprintf(`
export START_DATE=%s 
export END_DATE=%s
export FORECAST_DURATION=%d
eval $(prepvars)
rm -rf %s
dirprep --strict %s/%s %s`,
		startDate.Format("2006-01-02-15"),
		startDate.Add(time.Duration(durationHours)*time.Hour).Format("2006-01-02-15"),
		durationHours,
		targetDir,
		folders.TemplatesDir,
		name,
		targetDir,
	), folders.Rootdir, "")
}

func DirExists(directory string) bool {
	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return false
	}
	errors.Check(err)
	return info.IsDir()
}

func FileExists(file string) bool {
	info, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	errors.Check(err)
	return info.Mode().Type().IsRegular()
}
