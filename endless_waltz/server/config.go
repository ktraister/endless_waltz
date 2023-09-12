package main

// Configurations exported
type Configurations struct {
	Server ServerConfigurations
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Key       string
	Cert      string
	RandomURL string
	logLevel  string
	API_Key   string
}
