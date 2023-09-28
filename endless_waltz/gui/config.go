package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/user"
)

// Configurations exported
type Configurations struct {
	RandomURL   string
	ExchangeURL string
	LogLevel    string
	User        string
	Passwd      string
}

func fetchConfig() (Configurations, error) {
	var configuration Configurations
	//contents of temp config file
	contents := "randomURL: \"http://localhost:8090/api/otp\"\nexchangeURL: \"ws://localhost:8081/ws\"\nlogLevel: \"Debug\"\n"

	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current user: ", err)
		return configuration, err
	}

	// Get the user's home directory
	configDir := fmt.Sprintf("%s/.ew", currentUser.HomeDir)
	configFile := fmt.Sprintf("%s/config.yml", configDir)

	//check if directory exists and create if not
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Println("no config dir found, creating...")
		if err := os.Mkdir(configDir, os.ModePerm); err != nil {
			fmt.Println("Unable to create config home dir: ", err)
			return configuration, err
		}
	}

	//check if actual config file exists, create if not
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("no config file found, creating...")
		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println("Unable to create config file", err)
			return configuration, err
		}

		// Write contents to the file
		_, err = file.WriteString(contents)
		if err != nil {
			fmt.Println("Unable to write temp contents to config file", err)
			return configuration, err
		}
		file.Close()
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.ew/")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Viper:Error reading config file: ", err)
		return configuration, err
	}
	err = viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Println("Viper:Unable to decode into struct: ", err)
		return configuration, err
	}

	return configuration, nil
}
