package tag

import (
	"fmt"
	"strconv"

	"github.com/Sorrow446/go-mp4tag"
)

func writeMP4(path string, m *Metadata) error {
	t, err := mp4tag.Open(path)
	if err != nil {
		return fmt.Errorf("opening mp4: %w", err)
	}
	defer func() { _ = t.Close() }()

	tags := &mp4tag.MP4Tags{
		Title:       m.Title,
		Artist:      firstArtist(m.Artist),
		Album:       m.Album,
		AlbumArtist: m.AlbumArtist,
		CustomGenre: m.Genre,
		Composer:    m.Composer,
		Lyrics:      m.Lyrics,
		Date:        m.Date,
	}
	if tags.AlbumArtist == "" {
		tags.AlbumArtist = m.Artist
	}
	if n, err := strconv.Atoi(m.TrackNumber); err == nil {
		tags.TrackNumber = int16(n)
	}
	if n, err := strconv.Atoi(m.TotalTracks); err == nil {
		tags.TrackTotal = int16(n)
	}
	if n, err := strconv.Atoi(m.DiscNumber); err == nil {
		tags.DiscNumber = int16(n)
	}
	if n, err := strconv.Atoi(m.TotalDiscs); err == nil {
		tags.DiscTotal = int16(n)
	}

	if m.HasCover() {
		tags.Pictures = []*mp4tag.MP4Picture{{Data: m.Cover}}
	}

	if err := t.Write(tags, []string{}); err != nil {
		return fmt.Errorf("writing mp4 tags: %w", err)
	}
	return nil
}

func readMP4(path string) (*Metadata, error) {
	t, err := mp4tag.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening mp4: %w", err)
	}
	defer func() { _ = t.Close() }()

	tags, err := t.Read()
	if err != nil {
		return nil, fmt.Errorf("reading mp4 tags: %w", err)
	}
	m := &Metadata{
		Title:       tags.Title,
		Artist:      tags.Artist,
		Album:       tags.Album,
		AlbumArtist: tags.AlbumArtist,
		Genre:       tags.CustomGenre,
		Composer:    tags.Composer,
		Date:        tags.Date,
	}
	if tags.TrackNumber > 0 {
		m.TrackNumber = strconv.Itoa(int(tags.TrackNumber))
	}
	return m, nil
}
