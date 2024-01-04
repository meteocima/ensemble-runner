package mpiman_test

import (
	"testing"

	"github.com/meteocima/ensemble-runner/mpiman"
	"github.com/stretchr/testify/assert"
)

func TestMpiManager(t *testing.T) {

	t.Run("FindFreeNodes", func(t *testing.T) {
		nodes := mpiman.NewSlurmNodes()
		nodes.Nodes["a"] = true
		nodes.Nodes["b"] = false
		nodes.Nodes["c"] = true

		freeNodes, ok := nodes.FindFreeNodes(2)
		assert.True(t, ok)
		assert.Equal(t, mpiman.SlurmNodesList{"a", "c"}, freeNodes)
		for _, n := range nodes.Nodes {
			assert.False(t, n)
		}

		freeNodes2, ok2 := nodes.FindFreeNodes(1)
		assert.False(t, ok2)
		assert.Nil(t, freeNodes2)

		nodes.Dispose(freeNodes)
		freeNodes, ok = nodes.FindFreeNodes(1)
		assert.True(t, ok)
		assert.Equal(t, mpiman.SlurmNodesList{"a"}, freeNodes)

	})

	t.Run("ParseSlurmHosts", func(t *testing.T) {
		nodes, err := mpiman.ParseSlurmNodes("localhost")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"localhost"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("loc,al,host")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"al", "host", "loc"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes(",host")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"host"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("loc[host]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"lochost"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("un[loc,al,host]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"unal", "unhost", "unloc"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("[1-4,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"1", "2", "3", "4", "a"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("[001-4,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"001", "002", "003", "004", "a"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("[1-0004,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"1", "2", "3", "4", "a"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("[1-4]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"1", "2", "3", "4"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("h[1-4,a,b,06-8]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"h06", "h07", "h08", "h1", "h2", "h3", "h4", "ha", "hb"}, nodes.All())

		nodes, err = mpiman.ParseSlurmNodes("h[1-2],h2[06-8]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmNodesList{"h1", "h2", "h206", "h207", "h208"}, nodes.All())

		_, err = mpiman.ParseSlurmNodes("[-2]")
		assert.EqualError(t, err, "range start cannot be empty")

		_, err = mpiman.ParseSlurmNodes("[1-]")
		assert.EqualError(t, err, "range end cannot be empty")

		_, err = mpiman.ParseSlurmNodes("[a2-3]")
		assert.EqualError(t, err, "range start is not a number")

		_, err = mpiman.ParseSlurmNodes("[1-a2]")
		assert.EqualError(t, err, "range end is not a number")

		_, err = mpiman.ParseSlurmNodes("[]")
		assert.EqualError(t, err, "empty group")

		_, err = mpiman.ParseSlurmNodes("")
		assert.EqualError(t, err, "empty hosts list")
	})
}
