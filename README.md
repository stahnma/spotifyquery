# What is this?

This is a simple program to query your spotify app on your mac and give you a bit of what's going on in a json blob. 

It basically uses applescript and embeds that into the go project.

# Example

`./spotifyquery | jq`

```json
{
  "playing": true,
  "state": "playing",
  "sound_volume": 62,
  "shuffling": false,
  "shuffling_enabled": true,
  "repeating": false,
  "player_position_ms": 4609,
  "name": "Sushi and Coca-Cola",
  "artist": "St. Paul & The Broken Bones",
  "album": "Sushi and Coca-Cola",
  "album_artist": "St. Paul & The Broken Bones",
  "duration_ms": 157414,
  "track_number": 1,
  "disc_number": 1,
  "id": "spotify:track:5pGdWJIkt47ovJcBOKHl2S",
  "popularity": 55,
  "artwork_url": "https://i.scdn.co/image/ab67616d0000b273545d35897b007482838699fd",
  "collected_at": "2025-08-08T20:38:46.730808Z"
}


# License
MIT
