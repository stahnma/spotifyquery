package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//go:embed spotify_info.applescript
var spotifyScript string

type Output struct {
	Playing          bool    `json:"playing"`
	State            string  `json:"state"` // enum: not_running, stopped, paused, playing, unknown
	SoundVolume      *int    `json:"sound_volume,omitempty"`
	Shuffling        *bool   `json:"shuffling,omitempty"`
	ShufflingEnabled *bool   `json:"shuffling_enabled,omitempty"`
	Repeating        *bool   `json:"repeating,omitempty"`
	PlayerPositionMs *int64  `json:"player_position_ms,omitempty"`

	Name        *string `json:"name,omitempty"`
	Artist      *string `json:"artist,omitempty"`
	Album       *string `json:"album,omitempty"`
	AlbumArtist *string `json:"album_artist,omitempty"`
	DurationMs  *int64  `json:"duration_ms,omitempty"`
	TrackNumber *int    `json:"track_number,omitempty"`
	DiscNumber  *int    `json:"disc_number,omitempty"`
	ID          *string `json:"id,omitempty"`
	Popularity  *int    `json:"popularity,omitempty"`
	Starred     *bool   `json:"starred,omitempty"`
	ArtworkURL  *string `json:"artwork_url,omitempty"`

	CollectedAt string `json:"collected_at"`
	Error       string `json:"error,omitempty"`
}

func main() {
	raw, err := runOSA(spotifyScript)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// If osascript failed and gave nothing, emit an error payload.
	if err != nil && strings.TrimSpace(raw) == "" {
		emit(Output{CollectedAt: now, Error: err.Error()})
		return
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

	emit(out)
}

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

func emit(o Output) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.Encode(o)
}
