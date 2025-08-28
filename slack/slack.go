// Package slack provides Slack integration functionality for the spotifyquery application.
package slack

import (
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
)

// Service handles Slack operations
type Service struct {
	client    *slack.Client
	channelID string
}

// NewService creates a new Slack service
func NewService(botToken, channelID string) *Service {
	client := slack.New(botToken)
	return &Service{
		client:    client,
		channelID: channelID,
	}
}

// PostTrackInfo posts the track info with artist, song name, and both Spotify and YouTube links
// in a single message. Spotify links are clickable but don't unfurl, YouTube search links are formatted clearly.
func (s *Service) PostTrackInfo(artist, songName, spotifyURL string) error {
	// Create YouTube search URL
	searchQuery := fmt.Sprintf("%s %s", artist, songName)
	youtubeURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s",
		url.QueryEscape(searchQuery))

	// Post the main message with artist, song name, and both links
	// Disable unfurling to prevent URL previews, but links remain clickable
	message := fmt.Sprintf("%s - %s\n🎵 Spotify: %s\n🔍 YouTube: %s", artist, songName, spotifyURL, youtubeURL)

	_, _, err := s.client.PostMessage(
		s.channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionDisableLinkUnfurl(), // Prevent URL previews but links remain clickable
	)

	return err
}

// TestConnection tests the Slack connection
func (s *Service) TestConnection() error {
	_, err := s.client.AuthTest()
	return err
}
