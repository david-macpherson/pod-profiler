package config

import (
	"log"
	"os"
	"pod_profiler/pkg/api/config/env"
	"pod_profiler/pkg/api/defaults"

	"github.com/spf13/viper"
)

// This holds the config for the entire application
type Config struct {

	// The namespace the application is running in
	Namespace string `json:"namespace"`

	Deployments []string `json:"deployments"`

	*viper.Viper `json:"-"`
}

func Load(watchConfig bool) (*Config, error) {
	// Initialise an empty config
	config := &Config{}

	config.Viper = viper.New()
	config.Viper.SetDefault("deployments", []string{"bob"})
	config.Viper.SetDefault("namespace", defaults.NAMESPACE)
	config.Viper.BindEnv("namespace", "NAMESPACE")

	configName, configNameExists := os.LookupEnv("PROFILER_CONFIG_FILENAME")

	configDirs, err := env.GetConfigDirectories("pod-profiler", false)
	if err != nil {
		return nil, err
	}
	// Add the configuration directories to our search list
	for _, dir := range configDirs {
		config.Viper.AddConfigPath(dir)
	}

	// Attempt to parse our YAML configuration file if it exists in one of the directories
	if configNameExists {
		config.Viper.SetConfigName(configName)
	} else {
		config.Viper.SetConfigName("config")
	}

	config.Viper.SetConfigType("json")
	useConfigFile := true
	if err := config.Viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Default().Println("Configuration file not found, using configuration values from environment variables")
			useConfigFile = false
		} else {
			return nil, err
		}
	}

	// Watch the config file (if we found one) and ensure that we update our config when a change is detected
	if useConfigFile {

		// Check the watch config is true
		if watchConfig {

			// Start watching the config
			config.Viper.WatchConfig()
		}
	}

	// Unmarshal the config
	if err := config.Viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (config *Config) VarDump() {

	// Print our configuration values
	log.Default().Println("")
	log.Default().Println("-------------------------------")
	log.Default().Println("")
	log.Default().Printf("namespace:  %s\n", config.Namespace)
	log.Default().Printf("deployments:\n")

	for _, deployment := range config.Deployments {
		log.Default().Printf("\t%s\n", deployment)
	}

	log.Default().Println("")
	log.Default().Println("-------------------------------")
	log.Default().Println("")
}
