// Package cmd provides the command-line interface for the spotifyquery application.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stahnma/spotifyquery/config"
	"github.com/stahnma/spotifyquery/core"
)

var (
	cfgFile string
	post    bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spotifyquery",
	Short: "Query Spotify player information and optionally post to Slack",
	Long: `A command-line tool for querying Spotify player information on macOS using AppleScript.
It outputs detailed JSON data about the currently playing track, player state, and generates shareable URLs.
When the --post flag is used, it will also post the track information to a configured Slack channel.`,
	RunE: runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&post, "post", false, "post track information to Slack")

	// Bind flags to viper
	if err := viper.BindPFlag("post", rootCmd.PersistentFlags().Lookup("post")); err != nil {
		panic(fmt.Sprintf("failed to bind flag: %v", err))
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "config" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// runRoot executes the main logic
func runRoot(_ *cobra.Command, _ []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if we should post to Slack
	if post {
		if cfg.Slack.BotToken == "" {
			return fmt.Errorf("slack bot token is required. Set 'slack.bot_token' in config.yaml or SPOTIFYQUERY_SLACK_BOT_TOKEN")
		}
		if cfg.Slack.ChannelID == "" {
			return fmt.Errorf("slack channel ID is required. Set 'slack.channel_id' in config.yaml or SPOTIFYQUERY_SLACK_CHANNEL_ID")
		}
	}

	// Run the Spotify query logic
	return core.RunSpotifyQuery(post, cfg)
}
