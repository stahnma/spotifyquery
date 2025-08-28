# Spotify Query

A command-line tool for querying Spotify player information on macOS using AppleScript. It outputs detailed JSON data about the currently playing track, player state, and generates shareable URLs.

## Features

- **Real-time Spotify Data**: Get current track information, player state, and metadata
- **JSON Output**: Structured, colorized JSON output for easy parsing
- **Shareable URLs**: Automatically generates Spotify web player URLs
- **Slack Integration**: Optionally post track information to Slack channels
- **CLI Interface**: Modern command-line interface using Cobra and Viper

## Installation

```bash
go install github.com/stahnma/spotifyquery@latest
```

Or build from source:

```bash
git clone https://github.com/stahnma/spotifyquery.git
cd spotifyquery
go build
```

## Usage

### Basic Usage

```bash
# Query Spotify player information
./spotifyquery

# Output will be formatted JSON with track details
```

### Slack Integration

To post track information to Slack, use the `--post` flag:

```bash
# Post current track to Slack
./spotifyquery --post
```

**Note**: The tool posts only the shareable Spotify URI to Slack, which automatically unfurls to show a rich preview with track artwork, title, artist, and play button.

#### Configuration

Create a `config.yaml` file in your working directory:

```yaml
slack:
  bot_token: "xoxb-your-bot-token-here"
  channel_id: "C1234567890"
```

Or use environment variables:

```bash
export SPOTIFYQUERY_SLACK_BOT_TOKEN="xoxb-your-bot-token-here"
export SPOTIFYQUERY_SLACK_CHANNEL_ID="C1234567890"

./spotifyquery --post
```

#### Setting up Slack

1. Create a Slack app at [api.slack.com/apps](https://api.slack.com/apps)
2. Add the `chat:write` OAuth scope
3. Install the app to your workspace
4. Copy the Bot User OAuth Token (starts with `xoxb-`)
5. Get your channel ID from the channel URL or by right-clicking the channel name
6. Add the bot to your channel

### Command Line Options

```bash
./spotifyquery [flags]

Flags:
  --config string   config file (default is ./config.yaml)
  --post           post track information to Slack
  -h, --help       help for spotifyquery
```

## Output Format

The tool outputs JSON with the following structure:

```json
{
  "playing": false,
  "state": "paused",
  "sound_volume": 62,
  "shuffling": false,
  "shuffling_enabled": true,
  "repeating": false,
  "player_position_ms": 5835,
  "name": "Here Comes the Hotstepper",
  "artist": "Dr. Dog",
  "album": "Here Comes the Hotstepper",
  "album_artist": "Dr. Dog",
  "duration_ms": 286635,
  "track_number": 1,
  "disc_number": 1,
  "id": "spotify:track:3mkVy3Xrpeb2L4JBiwVSuC",
  "popularity": 35,
  "artwork_url": "https://i.scdn.co/image/ab67616d0000b273df9dc0ede1abeaaf01d2fb43",
  "share_url": "https://open.spotify.com/track/3mkVy3Xrpeb2L4JBiwVSuC",
  "collected_at": "2025-08-28T16:53:38.80735Z"
}
```

## Requirements

- macOS (uses AppleScript to query Spotify)
- Spotify desktop app installed and running
- Go 1.24.5 or later (for building from source)

## License

MIT License - see LICENSE file for details.
