package dirprep

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixtures = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot retrieve the source file path")
	} else {
		file = filepath.Dir(file)
	}

	return path.Join(file, "fixtures")
}()

func TestCopyExpanding(t *testing.T) {
	require.NoError(t, os.Setenv("cat", "42 cats"))
	defer func() {
		require.NoError(t, os.Unsetenv("cat"))
	}()

	src := strings.NewReader("song: $cat by row")
	var dest bytes.Buffer
	require.NoError(t, CopyExpanding(src, &dest, os.Getenv))
	assert.Equal(t, "song: 42 cats by row\n", dest.String())

	src = strings.NewReader("song: ${cat}byrow")
	dest.Reset()
	require.NoError(t, CopyExpanding(src, &dest, os.Getenv))
	assert.Equal(t, "song: 42 catsbyrow\n", dest.String())
}

func TestRecurseDir(t *testing.T) {
	cleanResult(t)

	os.Setenv("VAR1", "RAFMIC")

	var res []string
	_, err := RecurseDir(fixtures+"/readwrite", "/usr", func(src, dst string, d fs.DirEntry, mapping func(key string) string) error {
		if !d.IsDir() {
			res = append(res, fmt.Sprintf("%s -> %s\n", src, dst))
		}
		return nil
	}, os.Getenv)
	require.NoError(t, err)

	sort.Strings(res)
	assert.Equal(t, []string{
		fixtures + "/readwrite/dira/dirb/file${VAR1}1 -> /usr/dira/dirb/fileRAFMIC1\n",
		fixtures + "/readwrite/dira/dirb/file4 -> /usr/dira/dirb/file4\n",
		fixtures + "/readwrite/dira/dirc/file2 -> /usr/dira/dirc/file2\n",
		fixtures + "/readwrite/dira/file3 -> /usr/dira/file3\n",
	}, res)
}

func TestEnvExpander(t *testing.T) {
	cleanResult(t)

	require.NoError(t, os.Setenv("VAR1", "CIAO"))
	_, err := RecurseDir(fixtures+"/readwrite/dira", fixtures+"/result", EnvExpander, os.Getenv)
	require.NoError(t, err)

}

func TestLinks(t *testing.T) {
	cleanResult(t)
	require.NoError(t, os.Setenv("INVAR", "INDIR"))
	require.NoError(t, os.Setenv("OUTVAR", "OUTDIR"))
	perms, err := RecurseDir(fixtures+"/links", fixtures+"/result", EnvExpander, os.Getenv)
	require.NoError(t, err)
	require.NoError(t, ApplyPermissions(perms))

	l, err := os.Readlink(fixtures + "/result/dira")
	require.NoError(t, err)
	assert.Equal(t, "../readwrite/dira/", l)

	l, err = os.Readlink(fixtures + "/result/b")
	require.NoError(t, err)
	assert.Equal(t, "a", l)

	l, err = os.Readlink(fixtures + "/result/file3")
	require.NoError(t, err)
	assert.Equal(t, "../readwrite/dira/file3", l)

	l, err = os.Readlink(fixtures + "/result/OUTDIR")
	require.NoError(t, err)
	assert.Equal(t, "INDIR", l)

}

func cleanResult(t *testing.T) {
	syscall.Umask(0)

	err := os.Chmod(fixtures+"/result", 0777)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	} else {
		require.NoError(t, err)
		require.NoError(t, os.RemoveAll(fixtures+"/result"))
	}
}

func TestApplyPermissionsRODir(t *testing.T) {

	cleanResult(t)

	perms, err := RecurseDir(fixtures+"/readonly", fixtures+"/result", EnvExpander, os.Getenv)
	require.NoError(t, err)
	require.FileExists(t, fixtures+"/result/file1")
	err = ApplyPermissions(perms)
	require.NoError(t, err)

	infodst, err := os.Stat(fixtures + "/result/file1")
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0664), infodst.Mode())

}

func TestWODir(t *testing.T) {
	cleanResult(t)

	require.NoError(t, os.Chmod(fixtures+"/writeonly", 0333))
	defer os.Chmod(fixtures+"/writeonly", 0777)
	os.RemoveAll(fixtures + "/result")
	_, err := RecurseDir(fixtures+"/writeonly", fixtures+"/result", EnvExpander, os.Getenv)
	require.Error(t, err)
}
