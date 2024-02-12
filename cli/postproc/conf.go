package main

import (
	"os"

	"github.com/meteocima/ensemble-runner/errors"
	"gopkg.in/yaml.v3"
)

var Conf = struct {
	PostprocRules map[string]string `yaml:"PostprocRules"`
}{}

func ReadConf() {
	cfgFile := "./config.yaml"
	cfg := errors.CheckResult(os.ReadFile(cfgFile))
	errors.Check(yaml.Unmarshal(cfg, &Conf))

}
