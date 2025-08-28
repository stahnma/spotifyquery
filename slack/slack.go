package slack

import (
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

// PostTrackInfo posts just the shareable URI to Slack for automatic unfurling
func (s *Service) PostTrackInfo(shareURL string) error {
	// Post just the URL - Slack will automatically unfurl it with rich preview
	_, _, err := s.client.PostMessage(
		s.channelID,
		slack.MsgOptionText(shareURL, false),
	)

	return err
}

// TestConnection tests the Slack connection
func (s *Service) TestConnection() error {
	_, err := s.client.AuthTest()
	return err
}
