package main

// Configurations exported
type Configurations struct {
	Server       ServerConfigurations
}

// ServerConfigurations exported
type ServerConfigurations struct {
	MongoURI string
	UploadAPIKey string
}
