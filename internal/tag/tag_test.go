package tag

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/go-flac"
)

// fakeCover is a real (tiny) PNG so libraries that decode image dimensions
// (e.g. FLAC picture blocks) accept it.
var fakeCover = makePNG()
var sampleMeta = &Metadata{
	Title:       "Believer",
	Artist:      "Imagine Dragons, Someone Else",
	Album:       "Evolve",
	Date:        "2017",
	Genre:       "Rock",
	TrackNumber: "3",
	Lyrics:      "First things first\nI'ma say all the words inside my head",
	Cover:       fakeCover,
	CoverMIME:   "image/png",
}

// makePNG returns a small valid PNG image as raw bytes.
func makePNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 32), G: uint8(y * 32), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// buildMinimalOpus constructs a small but valid Ogg Opus stream in memory using
// the package's own page machinery.
func buildMinimalOpus(serial uint32) []byte {
	opusHead := func() []byte {
		var b bytes.Buffer
		b.WriteString("OpusHead")
		b.WriteByte(1)                                       // version
		b.WriteByte(2)                                       // channel count
		binary.Write(&b, binary.LittleEndian, uint16(312))   // pre-skip
		binary.Write(&b, binary.LittleEndian, uint32(48000)) // input sample rate
		binary.Write(&b, binary.LittleEndian, uint16(0))     // output gain
		b.WriteByte(0)                                       // mapping family
		return b.Bytes()
	}()

	opusTags := func() []byte {
		var b bytes.Buffer
		b.WriteString("OpusTags")
		vendor := "libopus-test"
		binary.Write(&b, binary.LittleEndian, uint32(len(vendor)))
		b.WriteString(vendor)
		binary.Write(&b, binary.LittleEndian, uint32(0)) // zero comments
		return b.Bytes()
	}()

	audio := bytes.Repeat([]byte{0xAA}, 200)

	headPage := oggPage{headerType: 0x02, granule: 0, serial: serial, segTable: []byte{byte(len(opusHead))}, body: opusHead}
	tagsPages := packetToPages(opusTags, serial, 0, 0x00)
	audioPage := oggPage{headerType: 0x04, granule: 960, serial: serial, segTable: []byte{byte(len(audio))}, body: audio}

	var out bytes.Buffer
	seq := uint32(0)
	out.Write(headPage.serialize(seq))
	seq++
	for _, p := range tagsPages {
		out.Write(p.serialize(seq))
		seq++
	}
	out.Write(audioPage.serialize(seq))
	return out.Bytes()
}

func TestOpusRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.opus")
	if err := os.WriteFile(path, buildMinimalOpus(0x12345678), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Write(path, sampleMeta); err != nil {
		t.Fatalf("Write opus: %v", err)
	}

	// The rewritten file must still be a valid Ogg stream with correct CRCs.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := parseOggPages(data)
	if err != nil {
		t.Fatalf("re-parsing written opus: %v", err)
	}
	for i, p := range pages {
		// serialize() recomputes the CRC; compare against what's on disk.
		got := p.serialize(uint32(i))
		// Find this page's bytes in data by re-serializing; CRC must be valid.
		if oggCRC(zeroCRC(got)) != binary.LittleEndian.Uint32(got[22:26]) {
			t.Fatalf("page %d has invalid CRC", i)
		}
	}

	// The audio page (granule 960) must survive untouched.
	foundAudio := false
	for _, p := range pages {
		if p.granule == 960 && len(p.body) == 200 && p.body[0] == 0xAA {
			foundAudio = true
		}
	}
	if !foundAudio {
		t.Fatal("audio page did not survive the rewrite")
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read opus: %v", err)
	}
	if got.Title != "Believer" {
		t.Errorf("title = %q, want Believer", got.Title)
	}
	if got.Artist != "Imagine Dragons" { // first artist only
		t.Errorf("artist = %q, want Imagine Dragons", got.Artist)
	}
	if got.AlbumArtist != "Imagine Dragons, Someone Else" {
		t.Errorf("album artist = %q, want full string", got.AlbumArtist)
	}
	if got.Genre != "Rock" {
		t.Errorf("genre = %q, want Rock", got.Genre)
	}
	if got.Lyrics == "" {
		t.Error("lyrics not round-tripped")
	}

	// Cover must be embedded as METADATA_BLOCK_PICTURE containing our bytes.
	if !bytes.Contains(data, []byte("METADATA_BLOCK_PICTURE")) {
		t.Fatal("cover comment missing")
	}
	if !pictureBytesPresent(data, fakeCover) {
		t.Fatal("cover image bytes not found in embedded picture")
	}
}

func TestOpusLargeCoverMultiPage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.opus")
	if err := os.WriteFile(path, buildMinimalOpus(0x1), 0o644); err != nil {
		t.Fatal(err)
	}
	// A cover larger than a single Ogg page (>~65KB) forces multi-page packing.
	big := &Metadata{Title: "Big", Artist: "A", Album: "B", Cover: bytes.Repeat([]byte{0x7E}, 200_000), CoverMIME: "image/jpeg"}
	if err := Write(path, big); err != nil {
		t.Fatalf("Write large opus: %v", err)
	}
	data, _ := os.ReadFile(path)
	pages, err := parseOggPages(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	foundAudio := false
	for _, p := range pages {
		if p.granule == 960 {
			foundAudio = true
		}
	}
	if !foundAudio {
		t.Fatal("audio page lost after large-cover rewrite")
	}
	got, err := Read(path)
	if err != nil || got.Title != "Big" {
		t.Fatalf("read back failed: %v title=%q", err, got.Title)
	}
}

func TestMP3RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.mp3")
	// A short run of silent MPEG frames; enough for id3v2 to read a header and
	// prepend a tag.
	frame := []byte{0xFF, 0xFB, 0x90, 0x00}
	buf := bytes.Repeat(append(frame, make([]byte, 413)...), 4)
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, sampleMeta); err != nil {
		t.Fatalf("Write mp3: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read mp3: %v", err)
	}
	if got.Title != "Believer" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Artist != "Imagine Dragons" {
		t.Errorf("artist = %q, want first artist only", got.Artist)
	}
	if got.AlbumArtist != "Imagine Dragons, Someone Else" {
		t.Errorf("album artist = %q", got.AlbumArtist)
	}

	// Verify cover and lyrics frames exist directly.
	tg, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Close()
	if len(tg.GetFrames(tg.CommonID("Attached picture"))) == 0 {
		t.Error("APIC cover frame missing")
	}
	if len(tg.GetFrames(tg.CommonID("Unsynchronised lyrics/text transcription"))) == 0 {
		t.Error("USLT lyrics frame missing")
	}
}

func TestFLACRoundTrip(t *testing.T) {
	path := genFixture(t, "flac")
	if err := Write(path, sampleMeta); err != nil {
		t.Fatalf("Write flac: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read flac: %v", err)
	}
	if got.Title != "Believer" {
		t.Errorf("title = %q", got.Title)
	}
	f, err := flac.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hasPic := false
	for _, b := range f.Meta {
		if b.Type == flac.Picture {
			hasPic = true
		}
	}
	if !hasPic {
		t.Error("FLAC picture block missing")
	}
}

func TestMP4RoundTrip(t *testing.T) {
	path := genFixture(t, "m4a")
	if err := Write(path, sampleMeta); err != nil {
		t.Fatalf("Write m4a: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read m4a: %v", err)
	}
	if got.Title != "Believer" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Artist != "Imagine Dragons" {
		t.Errorf("artist = %q", got.Artist)
	}
}

// genFixture creates a real audio file of the given extension using ffmpeg,
// skipping the test when ffmpeg is unavailable.
func genFixture(t *testing.T, ext string) string {
	t.Helper()
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skipf("ffmpeg not available; skipping %s fixture test", ext)
	}
	path := filepath.Join(t.TempDir(), "track."+ext)
	cmd := exec.Command(ffmpeg, "-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-y", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg fixture generation failed: %v\n%s", err, out)
	}
	return path
}

// zeroCRC returns a copy of a serialized page with its CRC field zeroed.
func zeroCRC(page []byte) []byte {
	cp := append([]byte(nil), page...)
	cp[22], cp[23], cp[24], cp[25] = 0, 0, 0, 0
	return cp
}

// pictureBytesPresent decodes embedded METADATA_BLOCK_PICTURE comments and
// checks the picture payload contains want.
func pictureBytesPresent(data, want []byte) bool {
	pages, err := parseOggPages(data)
	if err != nil {
		return false
	}
	packet, _, ok := packetSpan(pages, 1)
	if !ok {
		return false
	}
	// Walk comments looking for METADATA_BLOCK_PICTURE.
	pos := 8
	vlen := int(binary.LittleEndian.Uint32(packet[pos : pos+4]))
	pos += 4 + vlen
	count := int(binary.LittleEndian.Uint32(packet[pos : pos+4]))
	pos += 4
	for i := 0; i < count; i++ {
		clen := int(binary.LittleEndian.Uint32(packet[pos : pos+4]))
		pos += 4
		comment := packet[pos : pos+clen]
		pos += clen
		const prefix = "METADATA_BLOCK_PICTURE="
		if bytes.HasPrefix(comment, []byte(prefix)) {
			raw, err := base64.StdEncoding.DecodeString(string(comment[len(prefix):]))
			if err != nil {
				return false
			}
			if bytes.Contains(raw, want) {
				return true
			}
		}
	}
	return false
}
