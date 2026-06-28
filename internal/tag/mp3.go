package tag

import (
	"fmt"
	"strings"

	"github.com/bogem/id3v2/v2"
)

func writeMP3(path string, m *Metadata) error {
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("opening mp3: %w", err)
	}
	defer func() { _ = t.Close() }()

	enc := id3v2.EncodingUTF8

	if m.Title != "" {
		t.SetTitle(m.Title)
	}
	if m.Album != "" {
		t.SetAlbum(m.Album)
	}
	if m.Artist != "" {
		t.SetArtist(firstArtist(m.Artist))
	}

	albumArtist := m.AlbumArtist
	if albumArtist == "" {
		albumArtist = m.Artist
	}
	if albumArtist != "" {
		t.AddTextFrame(t.CommonID("Band/Orchestra/Accompaniment"), enc, albumArtist)
	}
	if m.Date != "" {
		t.SetYear(m.Date)
	}
	if m.Genre != "" {
		t.SetGenre(m.Genre)
	}
	if m.TrackNumber != "" {
		t.AddTextFrame(t.CommonID("Track number/Position in set"), enc, trackString(m.TrackNumber, m.TotalTracks))
	}
	if m.DiscNumber != "" {
		t.AddTextFrame(t.CommonID("Part of a set"), enc, trackString(m.DiscNumber, m.TotalDiscs))
	}
	if m.Composer != "" {
		t.AddTextFrame(t.CommonID("Composer"), enc, m.Composer)
	}

	if m.Lyrics != "" {
		t.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
			Encoding:          enc,
			Language:          "eng",
			ContentDescriptor: "",
			Lyrics:            m.Lyrics,
		})
	}

	if m.HasCover() {
		t.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    enc,
			MimeType:    coverMIME(m),
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     m.Cover,
		})
	}

	if err := t.Save(); err != nil {
		return fmt.Errorf("saving mp3 tags: %w", err)
	}
	return nil
}

func readMP3(path string) (*Metadata, error) {
	t, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("opening mp3: %w", err)
	}
	defer func() { _ = t.Close() }()

	m := &Metadata{
		Title:  t.Title(),
		Artist: t.Artist(),
		Album:  t.Album(),
		Genre:  t.Genre(),
		Date:   t.Year(),
	}
	if frames := t.GetFrames(t.CommonID("Band/Orchestra/Accompaniment")); len(frames) > 0 {
		if tf, ok := frames[0].(id3v2.TextFrame); ok {
			m.AlbumArtist = tf.Text
		}
	}
	if frames := t.GetFrames(t.CommonID("Track number/Position in set")); len(frames) > 0 {
		if tf, ok := frames[0].(id3v2.TextFrame); ok {
			m.TrackNumber, m.TotalTracks = splitSlashPair(tf.Text)
		}
	}
	if frames := t.GetFrames(t.CommonID("Part of a set")); len(frames) > 0 {
		if tf, ok := frames[0].(id3v2.TextFrame); ok {
			m.DiscNumber, m.TotalDiscs = splitSlashPair(tf.Text)
		}
	}
	for _, frame := range t.GetFrames("TXXX") {
		uf, ok := frame.(id3v2.UserDefinedTextFrame)
		if !ok {
			continue
		}
		switch strings.ToUpper(uf.Description) {
		case "TRACKTOTAL":
			if m.TotalTracks == "" {
				m.TotalTracks = uf.Value
			}
		case "DISCTOTAL":
			if m.TotalDiscs == "" {
				m.TotalDiscs = uf.Value
			}
		case "DISCNUMBER":
			if m.DiscNumber == "" {
				m.DiscNumber = uf.Value
			}
		}
	}
	return m, nil
}

// trackString renders "n/total" or just "n" for ID3 track/disc frames.
func trackString(n, total string) string {
	if total != "" {
		return fmt.Sprintf("%s/%s", n, total)
	}
	return n
}
