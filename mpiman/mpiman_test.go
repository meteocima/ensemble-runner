package mpiman_test

import (
	"testing"

	"github.com/meteocima/ensemble-runner/mpiman"
	"github.com/stretchr/testify/assert"
)

func TestMpiManager(t *testing.T) {
	t.Run("ParseSlurmHosts", func(t *testing.T) {
		hosts := mpiman.ParseSlurmHosts("localhost")
		assert.Equal(t, mpiman.SlurmHosts{"localhost"}, hosts)

		hosts = mpiman.ParseSlurmHosts("loc,al,host")
		assert.Equal(t, mpiman.SlurmHosts{"loc", "al", "host"}, hosts)

		assert.Panics(t, func() { mpiman.ParseSlurmHosts(",host") })

		hosts = mpiman.ParseSlurmHosts("loc[host]")
		assert.Equal(t, mpiman.SlurmHosts{"lochost"}, hosts)

		hosts = mpiman.ParseSlurmHosts("un[loc,al,host]")
		assert.Equal(t, mpiman.SlurmHosts{"unloc", "unal", "unhost"}, hosts)

		hosts = mpiman.ParseSlurmHosts("[1-4,a]")
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4", "a"}, hosts)

		hosts = mpiman.ParseSlurmHosts("[001-4,a]")
		assert.Equal(t, mpiman.SlurmHosts{"001", "002", "003", "004", "a"}, hosts)

		hosts = mpiman.ParseSlurmHosts("[1-0004,a]")
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4", "a"}, hosts)

		hosts = mpiman.ParseSlurmHosts("[1-4]")
		assert.Equal(t, mpiman.SlurmHosts{"1", "2", "3", "4"}, hosts)

		hosts = mpiman.ParseSlurmHosts("h[1-4,a,b,06-8]")
		assert.Equal(t, mpiman.SlurmHosts{"h1", "h2", "h3", "h4", "ha", "hb", "h06", "h07", "h08"}, hosts)

	})
}
