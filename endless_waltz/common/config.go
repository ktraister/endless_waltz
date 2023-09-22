package main

// Configurations exported
type Configurations struct {
	Server ServerConfigurations
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Key         string
	Cert        string
	RandomURL   string
	LogLevel    string
	Passwd      string
	ExchangeURL string
	User        string
}
