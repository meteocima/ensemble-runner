package wrfprocs_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/meteocima/ensemble-runner/wrfprocs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDAParser(t *testing.T) {
	t.Run("fixture", func(t *testing.T) {
		info, err := fs.Stat(fixtureFS, "rsl.out.wrfda.0000")
		require.NoError(t, err)
		require.Equal(t, "rsl.out.wrfda.0000", info.Name())
	})

	t.Run("ShowProgress", func(t *testing.T) {
		f, err := fixtureFS.Open("rsl.out.wrfita-filse-optim")
		require.NoError(t, err)
		defer f.Close()
		prgs := wrfprocs.ShowProgress(f,
			time.Date(2022, 11, 11, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 11, 13, 0, 0, 0, 0, time.UTC),
		)

		i := 1
		for p := range prgs {
			assert.Equal(t, i, p.Val)
			assert.NoError(t, p.Err)
			assert.False(t, p.Completed)
			if i == 100 {
				break
			}
			i++
		}
		p := <-prgs
		assert.Equal(t, 100, p.Val)
		assert.NoError(t, p.Err)
		assert.True(t, p.Completed)
		_, ok := <-prgs
		assert.False(t, ok)
	})

	t.Run("parse", func(t *testing.T) {
		f, err := fixtureFS.Open("simple.log")
		require.NoError(t, err)
		defer f.Close()
		p := wrfprocs.Parser{R: f}
		require.True(t, p.Read())
		assert.Equal(t, int64(3), p.Curr.Domain)
		assert.Equal(t, wrfprocs.CalcLine, p.Curr.Type)
		assert.Equal(t, 0.19767, p.Curr.Duration.Seconds())
		assert.Equal(t, "2022-12-06T02:00:00Z", p.Curr.Instant.Format(time.RFC3339))
		assert.Equal(t, float64(9), p.Curr.Timestep)

		require.True(t, p.Read())
		assert.Equal(t, wrfprocs.FileOutLine, p.Curr.Type)
		assert.Equal(t, int64(3), p.Curr.Domain)
		assert.Equal(t, 0.66925, p.Curr.Duration.Seconds())
		assert.True(t, p.Curr.Instant.IsZero())
		assert.Equal(t, "auxhist23_d03_2022-12-06_01:00:00", p.Curr.Filename)

		require.True(t, p.Read())
		assert.Equal(t, wrfprocs.FileInputLine, p.Curr.Type)
		assert.Equal(t, int64(3), p.Curr.Domain)
		assert.Equal(t, 2.18410, p.Curr.Duration.Seconds())
		assert.True(t, p.Curr.Instant.IsZero())
		assert.Equal(t, "wrfinput file (stream 0)", p.Curr.Filename)

		require.True(t, p.Read())
		assert.Equal(t, wrfprocs.FileInputLine, p.Curr.Type)
		assert.Equal(t, int64(1), p.Curr.Domain)
		assert.Equal(t, 0.25609, p.Curr.Duration.Seconds())
		assert.True(t, p.Curr.Instant.IsZero())
		assert.Equal(t, "lateral boundary", p.Curr.Filename)

		require.False(t, p.Read())
		require.NoError(t, p.Err)
	})

}
