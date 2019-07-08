package github

// Config for github API
type Config struct {
	RepoOwner             string `split_word:"true"`
	RepoName              string `split_word:"true"`
	DefaultCheckoutTarget string `split_word:"true"`
}

// Validate cfg after load
func (c *Config) Validate() error {
	return nil
}

// GetDefaultConfig return default config for github pkg
func GetDefaultConfig() Config {
	return Config{
		RepoOwner:             "makerdao",
		RepoName:              "dss-deploy-scripts",
		DefaultCheckoutTarget: "tags/qa-deploy",
	}
}
