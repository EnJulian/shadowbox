package tag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

const opusTagsMagic = "OpusTags"

// writeOpus rewrites the OpusTags comment header of an Ogg Opus file with the
// supplied metadata, cover art, and lyrics.
func writeOpus(path string, m *Metadata) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	pages, err := parseOggPages(data)
	if err != nil {
		return fmt.Errorf("parsing opus: %w", err)
	}
	if len(pages) < 2 {
		return errors.New("opus file has too few pages")
	}
	if !bytes.HasPrefix(pages[0].body, []byte("OpusHead")) {
		return errors.New("not an Ogg Opus stream (missing OpusHead)")
	}
	serial := pages[0].serial

	oldPacket, consumed, ok := packetSpan(pages, 1)
	if !ok {
		return errors.New("unable to isolate OpusTags packet for rewrite")
	}

	vendor := parseOpusVendor(oldPacket)
	newPacket := buildOpusTags(vendor, m)

	commentPages := packetToPages(newPacket, serial, 0, 0x00)

	newPages := make([]oggPage, 0, 1+len(commentPages)+len(pages)-(1+consumed))
	newPages = append(newPages, pages[0])
	newPages = append(newPages, commentPages...)
	newPages = append(newPages, pages[1+consumed:]...)

	var out bytes.Buffer
	for i, p := range newPages {
		out.Write(p.serialize(uint32(i)))
	}
	return os.WriteFile(path, out.Bytes(), 0o644)
}

// parseOpusVendor extracts the vendor string from an OpusTags packet, returning
// a default if it cannot be read.
func parseOpusVendor(packet []byte) string {
	if len(packet) < 12 || string(packet[:8]) != opusTagsMagic {
		return "shadowbox"
	}
	vlen := binary.LittleEndian.Uint32(packet[8:12])
	if 12+int(vlen) > len(packet) {
		return "shadowbox"
	}
	return string(packet[12 : 12+vlen])
}

// buildOpusTags constructs an OpusTags packet from metadata.
func buildOpusTags(vendor string, m *Metadata) []byte {
	var comments []string
	add := func(key, val string) {
		if val != "" {
			comments = append(comments, key+"="+val)
		}
	}
	add("TITLE", m.Title)
	add("ARTIST", firstArtist(m.Artist))
	add("ALBUM", m.Album)
	albumArtist := m.AlbumArtist
	if albumArtist == "" {
		albumArtist = m.Artist
	}
	add("ALBUMARTIST", albumArtist)
	add("DATE", m.Date)
	add("GENRE", m.Genre)
	add("TRACKNUMBER", m.TrackNumber)
	add("TRACKTOTAL", m.TotalTracks)
	add("DISCNUMBER", m.DiscNumber)
	add("DISCTOTAL", m.TotalDiscs)
	add("COMPOSER", m.Composer)
	add("PERFORMER", m.Performer)
	add("LYRICS", m.Lyrics)

	if m.HasCover() {
		block := buildFLACPictureBlock(m.Cover, coverMIME(m))
		add("METADATA_BLOCK_PICTURE", base64.StdEncoding.EncodeToString(block))
	}

	var buf bytes.Buffer
	buf.WriteString(opusTagsMagic)
	writeUint32LE(&buf, uint32(len(vendor)))
	buf.WriteString(vendor)
	writeUint32LE(&buf, uint32(len(comments)))
	for _, c := range comments {
		writeUint32LE(&buf, uint32(len(c)))
		buf.WriteString(c)
	}
	return buf.Bytes()
}

// buildFLACPictureBlock builds the binary METADATA_BLOCK_PICTURE payload used by
// both FLAC and Opus to embed cover art.
func buildFLACPictureBlock(pic []byte, mime string) []byte {
	var buf bytes.Buffer
	writeUint32BE(&buf, 3) // picture type: front cover
	writeUint32BE(&buf, uint32(len(mime)))
	buf.WriteString(mime)
	writeUint32BE(&buf, 0) // description length (empty)
	writeUint32BE(&buf, 0) // width
	writeUint32BE(&buf, 0) // height
	writeUint32BE(&buf, 0) // color depth
	writeUint32BE(&buf, 0) // colors used
	writeUint32BE(&buf, uint32(len(pic)))
	buf.Write(pic)
	return buf.Bytes()
}

func readOpus(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	pages, err := parseOggPages(data)
	if err != nil {
		return nil, fmt.Errorf("parsing opus: %w", err)
	}
	if len(pages) < 2 {
		return nil, errors.New("opus file has too few pages")
	}
	packet, _, ok := packetSpan(pages, 1)
	if !ok || len(packet) < 12 || string(packet[:8]) != opusTagsMagic {
		return &Metadata{}, nil
	}
	vlen := binary.LittleEndian.Uint32(packet[8:12])
	pos := 12 + int(vlen)
	if pos+4 > len(packet) {
		return &Metadata{}, nil
	}
	count := binary.LittleEndian.Uint32(packet[pos : pos+4])
	pos += 4

	m := &Metadata{}
	for i := uint32(0); i < count; i++ {
		if pos+4 > len(packet) {
			break
		}
		clen := int(binary.LittleEndian.Uint32(packet[pos : pos+4]))
		pos += 4
		if pos+clen > len(packet) {
			break
		}
		comment := string(packet[pos : pos+clen])
		pos += clen
		applyVorbisComment(m, comment)
	}
	return m, nil
}

// applyVorbisComment maps a "KEY=VALUE" comment onto the metadata struct.
func applyVorbisComment(m *Metadata, comment string) {
	eq := strings.IndexByte(comment, '=')
	if eq < 0 {
		return
	}
	key := strings.ToUpper(comment[:eq])
	val := comment[eq+1:]
	switch key {
	case "TITLE":
		m.Title = val
	case "ARTIST":
		m.Artist = val
	case "ALBUM":
		m.Album = val
	case "ALBUMARTIST":
		m.AlbumArtist = val
	case "DATE":
		m.Date = val
	case "GENRE":
		m.Genre = val
	case "TRACKNUMBER", "TRACK":
		m.TrackNumber = val
	case "TRACKTOTAL":
		m.TotalTracks = val
	case "DISCNUMBER", "DISC":
		m.DiscNumber = val
	case "DISCTOTAL":
		m.TotalDiscs = val
	case "COMPOSER":
		m.Composer = val
	case "PERFORMER":
		m.Performer = val
	case "LYRICS":
		m.Lyrics = val
	}
}

func writeUint32LE(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func writeUint32BE(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}
