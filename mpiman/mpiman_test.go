package mpiman_test

import (
	"testing"

	"github.com/meteocima/ensemble-runner/mpiman"
	"github.com/stretchr/testify/assert"
)

func TestMpiManager(t *testing.T) {
	t.Run("ParseSlurmHosts", func(t *testing.T) {
		hosts, err := mpiman.ParseHosts("localhost")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"localhost"}, hosts)

		hosts, err = mpiman.ParseHosts("loc,al,host")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"loc", "al", "host"}, hosts)

		hosts, err = mpiman.ParseHosts(",host")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"host"}, hosts)

		hosts, err = mpiman.ParseHosts("loc[host]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"lochost"}, hosts)

		hosts, err = mpiman.ParseHosts("un[loc,al,host]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"unloc", "unal", "unhost"}, hosts)

		hosts, err = mpiman.ParseHosts("[1-4,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4", "a"}, hosts)

		hosts, err = mpiman.ParseHosts("[001-4,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"001", "002", "003", "004", "a"}, hosts)

		hosts, err = mpiman.ParseHosts("[1-0004,a]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4", "a"}, hosts)

		hosts, err = mpiman.ParseHosts("[1-4]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4"}, hosts)

		hosts, err = mpiman.ParseHosts("h[1-4,a,b,06-8]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"h1", "h2", "h3", "h4", "ha", "hb", "h06", "h07", "h08"}, hosts)

		hosts, err = mpiman.ParseHosts("h[1-2],h2[06-8]")
		assert.NoError(t, err)
		assert.Equal(t, mpiman.SlurmHosts{"h1", "h2", "h206", "h207", "h208"}, hosts)

		_, err = mpiman.ParseHosts("[-2]")
		assert.EqualError(t, err, "range start cannot be empty")

		_, err = mpiman.ParseHosts("[1-]")
		assert.EqualError(t, err, "range end cannot be empty")

		_, err = mpiman.ParseHosts("[a2-3]")
		assert.EqualError(t, err, "range start is not a number")

		_, err = mpiman.ParseHosts("[1-a2]")
		assert.EqualError(t, err, "range end is not a number")

		_, err = mpiman.ParseHosts("[]")
		assert.EqualError(t, err, "empty group")

		_, err = mpiman.ParseHosts("")
		assert.EqualError(t, err, "empty hosts list")
	})
}
