package dirprep

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Transformer is a function that read the file or directory
// with src path, apply some kind of transformation and then
// save it to dst. argument d is the DirEntry for src.
type Transformer func(src, dst string, d fs.DirEntry, mapping func(key string) string) error

// isShellSpecialVar reports  whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// getShellName returns the name that begins the string and the number of bytes
// consumed to extract it. If the name is enclosed in {}, it's part of a ${}
// expansion and two more bytes are needed than the length of the name.
func getShellName(s string) (string, int) {

	switch {

	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		// Scan to closing brace
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2 // Bad syntax; eat "${}"
				}
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; eat "${"
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ {
	}
	return s[:i], i
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore.
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

// Expand replaces ${var} or $var in the string based on the mapping function.
// For example, expandEnv(s) is equivalent to os.Expand(s, os.Getenv).
func expandEnv(s string, mapping func(key string) string) string {
	var buf []byte
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			if name == "" && w > 0 {
				// Encountered invalid syntax; eat the
				// characters.
			} else if name == "" {
				// Valid syntax, but $ was not followed by a
				// name. Leave the dollar character untouched.
				buf = append(buf, s[j])
			} else {
				buf = append(buf, mapping(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return string(buf) + s[i:]
}

// EnvExpander is a Transformer function that
// does the following things:
//
//   - when src is a directory, expand with os.EnvExpand the
//     dst path, and then create a directory with the resulting path.
//
//   - when src is a file, expand with os.EnvExpand the
//     dst path, create it as a new file, and copy the content
//     from src using `CopyExpanding`
func EnvExpander(src, dst string, d fs.DirEntry, mapping func(key string) string) (err error) {
	if err != nil {
		return fmt.Errorf("EnvExpander: cannot read source file info %s: %w", src, err)
	}

	if d.IsDir() {
		// recreate directories
		err := os.MkdirAll(dst, fs.FileMode(0777))
		if err != nil {
			return fmt.Errorf("EnvExpander: cannot create target directory %s: %w", src, err)
		}
		return nil
	}

	if d.Type()&os.ModeSymlink == os.ModeSymlink {
		// recreate symbolic links

		linkDst, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("EnvExpander: cannot read source symlink %s: %w", src, err)
		}
		linkDst = expandEnv(linkDst, mapping)

		if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("EnvExpander: cannot delete existing symlink %s: %w", dst, err)
		}

		err = os.Symlink(linkDst, dst)
		if err != nil {
			return fmt.Errorf("EnvExpander: cannot create target symlink %s: %w", src, err)
		}
		return nil
	}

	// copy & expand regular files
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("EnvExpander: cannot open source file %s: %w", src, err)
	}
	defer func() {
		err = errors.Join(err, r.Close())
	}()

	w, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.FileMode(0777))
	if err != nil {
		return fmt.Errorf("EnvExpander: cannot create target file %s: %w", dst, err)
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()

	return CopyExpanding(r, w, mapping)
}

type Permissions map[string]fs.FileMode

// RecurseDir walks the content of `srcdir`, and for
// every file or directory found, it call the `tr`
// function passing the src path and the corresponding
// path under dstdir
func RecurseDir(srcdir, dstdir string, tr Transformer, mapping func(key string) string) (Permissions, error) {
	perm := Permissions{}
	err := filepath.WalkDir(srcdir, func(src string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(dstdir, strings.TrimPrefix(src, srcdir))
		dst = expandEnv(dst, mapping)

		i, err := d.Info()
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != os.ModeSymlink {
			// symbolic links permissions cannot be
			// changed and is always 0777
			perm[dst] = i.Mode()
		}
		return tr(src, dst, d, mapping)
	})
	if err != nil {
		return nil, fmt.Errorf("RecurseDir: cannot walk `%s` directory: %w", srcdir, err)

	}
	return perm, nil
}

// CopyExpanding read from `r` line by line,
// exe each line using `expandEnv`
// and write the result to `w`
func CopyExpanding(r io.Reader, w io.Writer, mapping func(key string) string) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		l := expandEnv(scanner.Text(), mapping)

		if _, err := w.Write([]byte(l + "\n")); err != nil {
			return fmt.Errorf("CopyExpanding: cannot write to target file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("CopyExpanding: cannot read from source file: %w", err)
	}

	return nil
}

func ApplyPermissions(perms Permissions) error {
	for file, perm := range perms {
		err := os.Chmod(file, perm)
		if err != nil {
			return fmt.Errorf("ApplyPermissions: cannot apply permissions to %s: %w", file, err)
		}
	}
	return nil
}

func RenderDirEnv(srcdir, dstdir string, mapping func(key string) string) error {
	perms, err := RecurseDir(srcdir, dstdir, EnvExpander, mapping)
	if err != nil {
		return fmt.Errorf("RenderDirEnv: %w", err)
	}
	err = ApplyPermissions(perms)
	if err != nil {
		return fmt.Errorf("RenderDirEnv: %w", err)
	}
	return nil
}
