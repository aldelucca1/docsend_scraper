package fs

// Config - The configuration information for connecting to a mongodb
// instance
type Config struct {
	outputPath string
}

// NewConfig - Creates a new Config with the default values
func NewConfig() *Config {
	return &Config{
		outputPath: "/tmp/",
	}
}

// WithOutputPath - Set the output path to use
func (c *Config) WithOutputPath(outputPath string) *Config {
	c.outputPath = outputPath
	return c
}
