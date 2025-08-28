// Package main provides a command-line tool for querying Spotify player information
// on macOS using AppleScript. It outputs detailed JSON data about the currently
// playing track, player state, and generates shareable URLs.
package main

import (
	"github.com/stahnma/spotifyquery/cmd"
)

func main() {
	cmd.Execute()
}
