-- spotify_info.applescript

on replaceText(findText, replText, sourceText)
	set AppleScript's text item delimiters to findText
	set ti to text items of (sourceText as text)
	set AppleScript's text item delimiters to replText
	set res to ti as text
	set AppleScript's text item delimiters to ""
	return res
end replaceText

on escTSV(s)
	try
		set t to s as text
	on error
		set t to ""
	end try
	set t to replaceText(return, " ", t)
	set t to replaceText(linefeed, " ", t)
	set t to replaceText(tab, " ", t)
	return t
end escTSV

try
	if (application "Spotify" is not running) then
		return "state" & tab & "not_running" & linefeed & "playing" & tab & "false"
	end if
on error
	return "state" & tab & "not_running" & linefeed & "playing" & tab & "false"
end try

set lf to linefeed
set sep to tab

tell application "Spotify"
	set out to ""

	-- Player state
	set ps to player state
	set out to out & "state" & sep & (ps as text) & lf
	set out to out & "playing" & sep & ((ps is playing) as text) & lf

	-- App props
	try
		set out to out & "sound_volume" & sep & (sound volume as text) & lf
	end try
	try
		set out to out & "shuffling" & sep & (shuffling as text) & lf
	end try
	try
		set out to out & "repeating" & sep & (repeating as text) & lf
	end try
	try
		set out to out & "shuffling_enabled" & sep & (shuffling enabled as text) & lf
	end try
	try
		set out to out & "player_position" & sep & (player position as text) & lf
	end try

	-- Bail early if nothing is loaded
	if ps is stopped then return out

	-- Track ref (guarded)
	try
		set t to current track
	on error
		return out
	end try

	-- Optional debug
	try
		set out to out & "current_track_class" & sep & (class of t as text) & lf
	end try

	-- Track fields
	try
		set out to out & "id" & sep & (my escTSV(id of t)) & lf
	end try
	try
		set out to out & "name" & sep & (my escTSV(name of t)) & lf
	end try
	try
		set out to out & "artist" & sep & (my escTSV(artist of t)) & lf
	end try
	try
		set out to out & "album" & sep & (my escTSV(album of t)) & lf
	end try
	try
		set out to out & "album_artist" & sep & (my escTSV(album artist of t)) & lf
	end try
	try
		set out to out & "duration" & sep & (duration of t as text) & lf
	end try
	try
		set out to out & "track_number" & sep & (track number of t as text) & lf
	end try
	try
		set out to out & "disc_number" & sep & (disc number of t as text) & lf
	end try
	try
		set out to out & "popularity" & sep & (popularity of t as text) & lf
	end try
	try
		set out to out & "starred" & sep & (starred of t as text) & lf
	end try
	try
		set out to out & "artwork_url" & sep & (my escTSV(artwork url of t)) & lf
	end try

	return out
end tell
