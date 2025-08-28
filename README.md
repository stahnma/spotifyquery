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
  "playing": true,
  "state": "playing",
  "name": "Song Title",
  "artist": "Artist Name",
  "album": "Album Name",
  "share_url": "https://open.spotify.com/track/123456",
  "collected_at": "2024-01-01T12:00:00.000000000Z"
}
```

## Requirements

- macOS (uses AppleScript to query Spotify)
- Spotify desktop app installed and running
- Go 1.24.5 or later (for building from source)

## License

MIT License - see LICENSE file for details.
