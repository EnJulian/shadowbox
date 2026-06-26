package tag

import (
	"fmt"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

func writeFLAC(path string, m *Metadata) error {
	f, err := flac.ParseFile(path)
	if err != nil {
		return fmt.Errorf("parsing flac: %w", err)
	}

	// Drop any existing comment/picture blocks so we write a clean set.
	kept := f.Meta[:0]
	for _, b := range f.Meta {
		if b.Type == flac.VorbisComment || b.Type == flac.Picture {
			continue
		}
		kept = append(kept, b)
	}
	f.Meta = kept

	cmt := flacvorbis.New()
	addFLAC := func(key, val string) {
		if val != "" {
			_ = cmt.Add(key, val)
		}
	}
	addFLAC(flacvorbis.FIELD_TITLE, m.Title)
	addFLAC(flacvorbis.FIELD_ARTIST, firstArtist(m.Artist))
	addFLAC(flacvorbis.FIELD_ALBUM, m.Album)
	albumArtist := m.AlbumArtist
	if albumArtist == "" {
		albumArtist = m.Artist
	}
	addFLAC("ALBUMARTIST", albumArtist)
	addFLAC(flacvorbis.FIELD_DATE, m.Date)
	addFLAC(flacvorbis.FIELD_GENRE, m.Genre)
	addFLAC(flacvorbis.FIELD_TRACKNUMBER, m.TrackNumber)
	addFLAC("TRACKTOTAL", m.TotalTracks)
	addFLAC("DISCNUMBER", m.DiscNumber)
	addFLAC("DISCTOTAL", m.TotalDiscs)
	addFLAC("COMPOSER", m.Composer)
	addFLAC("PERFORMER", m.Performer)
	addFLAC("LYRICS", m.Lyrics)

	cmtBlock := cmt.Marshal()
	f.Meta = append(f.Meta, &cmtBlock)

	if m.HasCover() {
		pic, err := flacpicture.NewFromImageData(
			flacpicture.PictureTypeFrontCover,
			"Front cover",
			m.Cover,
			coverMIME(m),
		)
		if err == nil {
			picBlock := pic.Marshal()
			f.Meta = append(f.Meta, &picBlock)
		}
	}

	if err := f.Save(path); err != nil {
		return fmt.Errorf("saving flac: %w", err)
	}
	return nil
}

func readFLAC(path string) (*Metadata, error) {
	f, err := flac.ParseFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing flac: %w", err)
	}
	m := &Metadata{}
	for _, b := range f.Meta {
		if b.Type != flac.VorbisComment {
			continue
		}
		cmt, err := flacvorbis.ParseFromMetaDataBlock(*b)
		if err != nil {
			continue
		}
		get := func(key string) string {
			if vals, err := cmt.Get(key); err == nil && len(vals) > 0 {
				return vals[0]
			}
			return ""
		}
		m.Title = get(flacvorbis.FIELD_TITLE)
		m.Artist = get(flacvorbis.FIELD_ARTIST)
		m.Album = get(flacvorbis.FIELD_ALBUM)
		m.AlbumArtist = get("ALBUMARTIST")
		m.Date = get(flacvorbis.FIELD_DATE)
		m.Genre = get(flacvorbis.FIELD_GENRE)
		m.TrackNumber = get(flacvorbis.FIELD_TRACKNUMBER)
	}
	return m, nil
}
