package core

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stahnma/spotifyquery/config"
	"github.com/stahnma/spotifyquery/slack"
	"golang.org/x/term"
)

//go:embed spotify_info.applescript
var spotifyScript string

// Output represents the complete state and metadata of the Spotify player.
// It includes player state, current track information, and derived data like
// shareable URLs. All pointer fields use omitempty to exclude null values
// from JSON output when the data is unavailable.
type Output struct {
	// Player state fields
	Playing          bool   `json:"playing"`                      // Whether audio is currently playing
	State            string `json:"state"`                        // Player state: not_running, stopped, paused, playing, unknown
	SoundVolume      *int   `json:"sound_volume,omitempty"`       // Volume level (0-100)
	Shuffling        *bool  `json:"shuffling,omitempty"`          // Whether shuffle mode is active
	ShufflingEnabled *bool  `json:"shuffling_enabled,omitempty"`  // Whether shuffle mode is available
	Repeating        *bool  `json:"repeating,omitempty"`          // Whether repeat mode is active
	PlayerPositionMs *int64 `json:"player_position_ms,omitempty"` // Current position in milliseconds

	// Track metadata fields
	Name        *string `json:"name,omitempty"`         // Track title
	Artist      *string `json:"artist,omitempty"`       // Primary artist name
	Album       *string `json:"album,omitempty"`        // Album title
	AlbumArtist *string `json:"album_artist,omitempty"` // Album artist (may differ from track artist)
	DurationMs  *int64  `json:"duration_ms,omitempty"`  // Total track duration in milliseconds
	TrackNumber *int    `json:"track_number,omitempty"` // Track number within album
	DiscNumber  *int    `json:"disc_number,omitempty"`  // Disc number for multi-disc albums
	ID          *string `json:"id,omitempty"`           // Spotify URI or track ID
	Popularity  *int    `json:"popularity,omitempty"`   // Spotify popularity score (0-100)
	Starred     *bool   `json:"starred,omitempty"`      // Whether track is starred/liked
	ArtworkURL  *string `json:"artwork_url,omitempty"`  // Album artwork image URL

	// Derived fields
	ShareURL *string `json:"share_url,omitempty"` // HTTPS shareable URL derived from Spotify ID

	// Metadata fields
	CollectedAt string `json:"collected_at"`    // ISO 8601 timestamp when data was collected
	Error       string `json:"error,omitempty"` // Error message if collection failed
}

// RunSpotifyQuery executes the main Spotify query logic
func RunSpotifyQuery(postToSlack bool, cfg *config.Config) error {
	raw, err := runOSA(spotifyScript)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// If osascript failed and gave nothing, emit an error payload.
	if err != nil && strings.TrimSpace(raw) == "" {
		emit(Output{CollectedAt: now, Error: err.Error()})
		return err
	}

	kv := parseTSV(raw)
	out := Output{CollectedAt: now}

	if e := strings.TrimSpace(kv["error"]); e != "" {
		out.Error = e
	}

	state := strings.TrimSpace(kv["state"])
	if state == "" {
		state = "unknown"
	}
	out.State = state

	switch state {
	case "not_running", "stopped":
		out.Playing = false
	case "playing":
		out.Playing = true
	case "paused":
		out.Playing = false
	default:
		if p := strings.TrimSpace(kv["playing"]); p != "" {
			out.Playing = (p == "true")
		}
	}

	// App props
	if s := kv["sound_volume"]; s != "" {
		if i, err := strconv.Atoi(s); err == nil {
			out.SoundVolume = &i
		}
	}
	if s := kv["shuffling"]; s != "" {
		b := (s == "true")
		out.Shuffling = &b
	}
	if s := kv["repeating"]; s != "" {
		b := (s == "true")
		out.Repeating = &b
	}
	if s := kv["shuffling_enabled"]; s != "" {
		b := (s == "true")
		out.ShufflingEnabled = &b
	}
	if s := kv["player_position"]; s != "" {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			ms := int64(f * 1000.0)
			out.PlayerPositionMs = &ms
		}
	}

	// Track props
	setStr := func(dst **string, key string) {
		if v := strings.TrimSpace(kv[key]); v != "" {
			*dst = &v
		}
	}
	setInt := func(dst **int, key string) {
		if v := strings.TrimSpace(kv[key]); v != "" {
			if i, err := strconv.Atoi(v); err == nil {
				*dst = &i
			}
		}
	}
	setI64 := func(dst **int64, key string) {
		if v := strings.TrimSpace(kv[key]); v != "" {
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				*dst = &i
			}
		}
	}
	setBool := func(dst **bool, key string) {
		if v := strings.TrimSpace(kv[key]); v != "" {
			b := strings.EqualFold(v, "true")
			*dst = &b
		}
	}

	setStr(&out.Name, "name")
	setStr(&out.Artist, "artist")
	setStr(&out.Album, "album")
	setStr(&out.AlbumArtist, "album_artist")
	setStr(&out.ID, "id")
	setStr(&out.ArtworkURL, "artwork_url")

	setI64(&out.DurationMs, "duration")
	setInt(&out.TrackNumber, "track_number")
	setInt(&out.DiscNumber, "disc_number")
	setInt(&out.Popularity, "popularity")
	setBool(&out.Starred, "starred")

	// Derive shareable URL from the ID/URI when possible.
	if out.ID != nil {
		if u, ok := spotifyShareURL(*out.ID); ok {
			out.ShareURL = &u
		}
	}

	// Post to Slack if requested and we have track information
	if postToSlack && out.Name != nil && out.Artist != nil && out.ShareURL != nil {
		slackService := slack.NewService(cfg.Slack.BotToken, cfg.Slack.ChannelID)

		if err := slackService.PostTrackInfo(*out.ShareURL); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to post to Slack: %v\n", err)
		}
	}

	emit(out)
	return nil
}

// runOSA executes an AppleScript using the osascript command.
// It pipes the script content to osascript's stdin and captures both
// stdout and stderr. Returns the stdout content and any execution errors.
func runOSA(script string) (string, error) {
	cmd := exec.Command("osascript", "-")
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// parseTSV parses tab-separated values from AppleScript output.
// Each line should contain a key-value pair separated by a tab character.
// Empty lines are ignored. Returns a map of key-value pairs.
func parseTSV(s string) map[string]string {
	m := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = parts[1]
		}
	}
	return m
}

// spotifyShareURL converts various Spotify identifier formats to shareable HTTPS URLs.
//
// Supported input formats:
//   - spotify:<type>:<id> (e.g., "spotify:track:1BpMw2vf4sWnFXy6liC5tD")
//   - http://open.spotify.com/<type>/<id> (normalized to https)
//   - https://open.spotify.com/<type>/<id> (returned as-is)
//
// Local files (spotify:local:*) are not supported since they don't have web URLs.
// Returns the shareable URL and true on success, or empty string and false on failure.
func spotifyShareURL(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}

	// Already a web URL? Normalize to https.
	if strings.HasPrefix(s, "https://open.spotify.com/") {
		return s, true
	}
	if strings.HasPrefix(s, "http://open.spotify.com/") {
		return "https://open.spotify.com/" + strings.TrimPrefix(s, "http://open.spotify.com/"), true
	}

	// spotify: URIs
	if strings.HasPrefix(s, "spotify:local:") {
		// Local files aren't hosted on Spotify's web player.
		return "", false
	}
	if strings.HasPrefix(s, "spotify:") {
		parts := strings.Split(s, ":")
		// Expect at least spotify:<type>:<id>
		if len(parts) >= 3 {
			typ := parts[1]
			id := parts[2]
			if typ != "" && id != "" {
				return "https://open.spotify.com/" + typ + "/" + id, true
			}
		}
	}

	// Bare IDs aren't convertible without knowing the type.
	return "", false
}

// ANSI color codes for JSON syntax highlighting
const (
	colorReset  = "\x1b[0m"
	colorKey    = "\x1b[94m" // bright blue (lighter for dark terminals)
	colorString = "\x1b[32m" // green
	colorNumber = "\x1b[36m" // cyan
	colorBool   = "\x1b[33m" // yellow
	colorNull   = "\x1b[90m" // gray
	colorBrace  = "\x1b[35m" // magenta
)

// isTTY determines if the given file descriptor represents a terminal that supports colors.
// It checks if the fd is a terminal using golang.org/x/term, then validates color support
// through environment variables. Respects the NO_COLOR environment variable to disable colors.
//
// Color support is determined by checking:
//   - COLORTERM environment variable (if set, colors are supported)
//   - TERM environment variable for common color-supporting values
//   - NO_COLOR environment variable (if set, colors are disabled)
func isTTY(fd int) bool {
	// Check if it's a terminal
	if !term.IsTerminal(fd) {
		return false
	}

	// Additional check for color support
	termEnv := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")
	noColor := os.Getenv("NO_COLOR")

	// Don't colorize if NO_COLOR is set
	if noColor != "" {
		return false
	}

	// Colorize if COLORTERM is set or TERM suggests color support
	return colorTerm != "" ||
		strings.Contains(termEnv, "color") ||
		strings.Contains(termEnv, "256") ||
		termEnv == "xterm" ||
		termEnv == "screen"
}

// colorizeJSON applies ANSI color codes to JSON output for terminal display.
// Colors are only applied when outputting to a color-supporting TTY.
//
// Color scheme:
//   - Keys: bright blue
//   - String values: green
//   - Numbers: cyan
//   - Booleans: yellow
//   - Null values: gray
//   - Structural elements (braces, brackets, commas): magenta
//
// The function uses carefully ordered regex patterns to avoid conflicts
// between different JSON element types.
func colorizeJSON(data []byte) []byte {
	isTTYResult := isTTY(int(os.Stdout.Fd()))

	if !isTTYResult {
		return data
	}

	s := string(data)

	// Apply colors in a careful order to avoid conflicts
	// 1. First, color the structural elements
	braceRegex := regexp.MustCompile(`([{}[\],])`)
	s = braceRegex.ReplaceAllString(s, colorBrace+`$1`+colorReset)

	// 2. Then color keys (quoted strings followed by colon)
	keyRegex := regexp.MustCompile(`"([^"]+)"\s*:`)
	s = keyRegex.ReplaceAllString(s, colorKey+`"$1"`+colorReset+`:`)

	// 3. Color boolean values - match exactly true/false as values
	boolRegex := regexp.MustCompile(`:\s*(true|false)\s*([,}])`)
	s = boolRegex.ReplaceAllString(s, `: `+colorBool+`$1`+colorReset+`$2`)

	// 4. Color null values
	nullRegex := regexp.MustCompile(`:\s*(null)\s*([,}])`)
	s = nullRegex.ReplaceAllString(s, `: `+colorNull+`$1`+colorReset+`$2`)

	// 5. Color numbers - match exactly numbers as values
	numberRegex := regexp.MustCompile(`:\s*(-?\d+(?:\.\d+)?)\s*([,}])`)
	s = numberRegex.ReplaceAllString(s, `: `+colorNumber+`$1`+colorReset+`$2`)

	// 6. Finally, color string values (everything else in quotes)
	stringRegex := regexp.MustCompile(`:\s*"([^"]*)"`)
	s = stringRegex.ReplaceAllString(s, `: `+colorString+`"$1"`+colorReset)

	return []byte(s)
}

// emit outputs the Spotify player data as formatted JSON.
// The output is pretty-printed with 2-space indentation and includes
// color syntax highlighting when outputting to a color-supporting terminal.
// When piped to files or non-color terminals, outputs plain JSON.
func emit(o Output) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(o)

	// Get the JSON bytes and apply coloring if outputting to a TTY
	jsonBytes := buf.Bytes()
	coloredBytes := colorizeJSON(jsonBytes)

	// Try using fmt.Print instead of os.Stdout.Write
	fmt.Print(string(coloredBytes))
}
