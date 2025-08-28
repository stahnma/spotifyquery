package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Slack SlackConfig `mapstructure:"slack"`
}

// SlackConfig holds Slack-specific configuration
type SlackConfig struct {
	BotToken  string `mapstructure:"bot_token"`
	ChannelID string `mapstructure:"channel_id"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SPOTIFYQUERY")

	// Bind environment variables
	viper.BindEnv("slack.bot_token", "SPOTIFYQUERY_SLACK_BOT_TOKEN")
	viper.BindEnv("slack.channel_id", "SPOTIFYQUERY_SLACK_CHANNEL_ID")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if no config file is found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
