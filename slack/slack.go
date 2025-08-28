package slack

import (
	"fmt"
	"net/url"
	"strings"

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

// PostTrackInfo posts the track info with artist and song name, then adds links in a threaded comment
func (s *Service) PostTrackInfo(artist, songName, spotifyURL string) error {
	// Post the main message with just artist and song name
	message := fmt.Sprintf("%s - %s", artist, songName)
	
	_, timestamp, err := s.client.PostMessage(
		s.channelID,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return err
	}
	
	// Create YouTube search URL
	searchQuery := fmt.Sprintf("%s %s", artist, songName)
	youtubeURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", 
		url.QueryEscape(searchQuery))
	
	// Post links as a threaded reply
	linksMessage := fmt.Sprintf("Spotify: %s\nYouTube: %s", spotifyURL, youtubeURL)
	
	_, _, err = s.client.PostMessage(
		s.channelID,
		slack.MsgOptionText(linksMessage, false),
		slack.MsgOptionTS(timestamp), // This makes it a threaded reply
	)

	return err
}

// extractSpotifyTrackID extracts the track ID from a Spotify URL
func extractSpotifyTrackID(spotifyURL string) string {
	// Handle both https://open.spotify.com/track/ID and spotify:track:ID formats
	if strings.Contains(spotifyURL, "open.spotify.com/track/") {
		parts := strings.Split(spotifyURL, "/track/")
		if len(parts) > 1 {
			// Remove any query parameters
			trackID := strings.Split(parts[1], "?")[0]
			return trackID
		}
	} else if strings.HasPrefix(spotifyURL, "spotify:track:") {
		return strings.TrimPrefix(spotifyURL, "spotify:track:")
	}
	return spotifyURL // fallback to original URL
}

// TestConnection tests the Slack connection
func (s *Service) TestConnection() error {
	_, err := s.client.AuthTest()
	return err
}
