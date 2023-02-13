package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigurationFile struct {
	Address     string
	LoopSeconds uint `yaml:"loopSeconds"`
	Unseal      ConfigurationFileUnseal
}

type ConfigurationFileUnseal struct {
	Key  string
	Env  string
	File string
}

var (
	flagShowHelp              bool
	flagVerboseOutput         bool
	flagConfigurationFilename string
	flagAddress               string
	flagLoopSeconds           uint
	flagUnsealKey             string
	flagUnsealEnv             string
	flagUnsealFile            string
)

func (cf *ConfigurationFile) getUnsealKey() (string, error) {
	if cf.Unseal.Key != "" {
		log.Println("Using passed unseal key")

		return cf.Unseal.Key, nil
	}

	if cf.Unseal.Env != "" {
		value, present := os.LookupEnv(cf.Unseal.Env)

		if !present {
			return "", fmt.Errorf("%s %s", "Unseal environment var is not set:", cf.Unseal.Env)
		}

		log.Printf("Using environment var %s for unseal key\n", cf.Unseal.Env)

		return value, nil
	}

	if cf.Unseal.File != "" {
		unsealKeyBody, err := os.ReadFile(cf.Unseal.File)

		if err != nil {
			return "", fmt.Errorf("%s %s", "Error reading unseal key file:", err)
		}

		log.Printf("Using file %s for unseal key\n", cf.Unseal.File)

		return strings.TrimSpace(string(unsealKeyBody)), nil
	}

	return "", fmt.Errorf("%s", "No unseal key provided")
}

func initConfiguration() {
	flag.BoolVar(&flagShowHelp, "help", false, "Show help")
	flag.BoolVar(&flagVerboseOutput, "verbose", false, "Show verbose output")
	flag.StringVar(&flagConfigurationFilename, "config", "conf/configuration.yaml", "Path to the configuration file to load")
	flag.StringVar(&flagAddress, "address", "", "The address of the Vault server")
	flag.UintVar(&flagLoopSeconds, "loop-seconds", 0, "The time to wait between loop iterations for unseal attempts")
	flag.StringVar(&flagUnsealKey, "unseal-key", "", "The unseal key")
	flag.StringVar(&flagUnsealEnv, "unseal-env", "", "The environmental variable which contains the unseal key")
	flag.StringVar(&flagUnsealFile, "unseal-file", "", "The path to the file which contains the unseal key")
}

func processConfiguration() (*ConfigurationFile, error) {
	flag.Parse()

	if flagShowHelp {
		flag.Usage()

		os.Exit(0)
	}

	configurationFileBody, err := os.ReadFile(flagConfigurationFilename)

	if err != nil {
		return nil, fmt.Errorf("%s %s", "Error reading configuration file:", err)
	}

	var cf ConfigurationFile

	err = yaml.Unmarshal(configurationFileBody, &cf)

	if err != nil {
		return nil, fmt.Errorf("%s %s", "Error decoding configuration YAML:", err)
	}

	if flagAddress != "" {
		cf.Address = flagAddress
	}

	if flagLoopSeconds != 0 {
		cf.LoopSeconds = flagLoopSeconds
	}

	// Blanking out of unseal vars below is so that *any* flags passed in take
	// precedence over the configuration file, and flag checks are in the order
	// of least- to most-important flags

	if flagUnsealFile != "" {
		cf.Unseal.File = flagUnsealFile
		cf.Unseal.Env = ""
		cf.Unseal.Key = ""
	}

	if flagUnsealEnv != "" {
		cf.Unseal.File = ""
		cf.Unseal.Env = flagUnsealEnv
		cf.Unseal.Key = ""
	}

	if flagUnsealKey != "" {
		cf.Unseal.File = ""
		cf.Unseal.Env = ""
		cf.Unseal.Key = flagUnsealKey
	}

	return &cf, nil
}
