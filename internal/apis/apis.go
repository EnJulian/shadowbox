// Package apis and its subpackages provide thin HTTP clients for the external
// metadata services Shadowbox uses: Spotify, iTunes, Last.fm, and Genius.
package apis

// TrackMeta is the normalised metadata Shadowbox collects for a track. Numeric
// fields are kept as strings so they map directly onto tag fields.
type TrackMeta struct {
	Title       string
	Artist      string // full, comma-joined list of artists
	Album       string
	ReleaseDate string // YYYY or full date
	Genre       string
	TrackNumber string
	TotalTracks string
	DiscNumber  string
	TotalDiscs  string
	Composer    string
	Performer   string
	CoverURL    string
}

// Year returns just the four-digit year from ReleaseDate when possible.
func (m *TrackMeta) Year() string {
	if len(m.ReleaseDate) >= 4 {
		return m.ReleaseDate[:4]
	}
	return m.ReleaseDate
}
