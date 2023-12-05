package wrfprocs_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/meteocima/ensemble-runner/wrfprocs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetgrid(t *testing.T) {
	t.Run("fixture", func(t *testing.T) {
		info, err := fs.Stat(fixtureFS, "metgrid.log")
		require.NoError(t, err)
		require.Equal(t, "metgrid.log", info.Name())
	})

	t.Run("ShowProgress", func(t *testing.T) {
		f, err := fixtureFS.Open("metgrid.log")
		require.NoError(t, err)
		defer f.Close()
		prgs := wrfprocs.ShowMetgridProgress(f,
			time.Date(2020, 1, 9, 12, 0, 0, 0, time.UTC),
			time.Date(2020, 1, 10, 12, 0, 0, 0, time.UTC),
		)

		last := 0
		for p := range prgs {
			//fmt.Println(p)
			assert.Greater(t, p.Val, last)
			last = p.Val
			assert.NoError(t, p.Err)
			assert.False(t, p.Completed)
			if last == 100 {
				break
			}
		}
		p := <-prgs
		//fmt.Println(p)

		assert.Equal(t, 100, p.Val)
		assert.NoError(t, p.Err)
		assert.True(t, p.Completed)
		_, ok := <-prgs
		assert.False(t, ok)
	})
}
