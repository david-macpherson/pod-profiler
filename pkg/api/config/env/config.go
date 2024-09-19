package env

import (
	"os"
	"path"
	"runtime"
)

// GetConfigDirectories returns a set of candidate configuration directory locations with the specified application name appended.
// The list of entries is customised based on the platform.
func GetConfigDirectories(application string, includeCwd bool) ([]string, error) {
	
	// Determine whether we are including the current working directory in our list
	dirs := []string{}
	if includeCwd {
		
		// Attempt to retrieve the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		
		// Add the current working directory to our list
		dirs = append(dirs, cwd)
	}
	
	// Attempt to retrieve the user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	
	// Add the user config directory to our list
	dirs = append(dirs, path.Join(configDir, application))
	
	// Add any platform-specific directories to our list
	if runtime.GOOS == "linux" {
		dirs = append(dirs, path.Join("/etc", application))
	}
	
	return dirs, nil
}
