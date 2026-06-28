package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const albumHTMLFixture = `<!DOCTYPE html><html><body>
<h2>Test Game Soundtrack</h2>
<p>Year: <b>2024</b><br>
Published by: <a href="/publisher/test">Test Publisher</a><br>
Album type: <b><a href="/game-soundtracks/ost">Soundtrack</a></b><br>
Number of Files: <b>3</b></p>
<div class="albumImage">
<a href="https://nu.vgmtreasurechest.com/soundtracks/test/00%20Front.jpg">
<img src="https://nu.vgmtreasurechest.com/soundtracks/test/thumbs/00%20Front.jpg">
</a></div>
<table>
<tr><td><a href="/game-soundtracks/album/test-album/01.%20Opening.mp3">01. Opening</a></td></tr>
<tr><td><a href="/game-soundtracks/album/test-album/02.%20Battle.mp3">02. Battle</a></td></tr>
<tr><td><a href="/game-soundtracks/album/test-album/01.%20Opening.mp3">duplicate</a></td></tr>
<tr><td><a href="/game-soundtracks/album/test-album/03.%20Ending.flac">03. Ending</a></td></tr>
<tr><td><a href="/other/page">skip</a></td></tr>
</table>
</body></html>`

const trackHTMLFixture = `<!DOCTYPE html><html><body>
<p>Song name: <b>Opening Theme</b></p>
<audio id="audio" src="https://nu.vgmtreasurechest.com/soundtracks/test/01-opening.mp3"></audio>
<a href="https://nu.vgmtreasurechest.com/soundtracks/test/01-opening.flac">Download FLAC</a>
<a href="https://nu.vgmtreasurechest.com/soundtracks/test/01-opening.mp3">Download MP3</a>
</body></html>`

func TestScrapeKHInsiderAlbum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, albumHTMLFixture)
	}))
	defer srv.Close()

	tracks, err := scrapeKHInsiderAlbum(context.Background(), srv.URL+"/game-soundtracks/album/test-album")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3", len(tracks))
	}
	if tracks[0].Index != 1 || tracks[0].Title != "Opening" {
		t.Errorf("track 1 = %+v", tracks[0])
	}
	if tracks[1].Title != "Battle" {
		t.Errorf("track 2 title = %q", tracks[1].Title)
	}
	if !strings.HasSuffix(tracks[2].PageURL, "03.%20Ending.flac") {
		t.Errorf("track 3 url = %q", tracks[2].PageURL)
	}
}

func TestResolveKHInsiderTrackPreferredFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, trackHTMLFixture)
	}))
	defer srv.Close()

	c := &khClient{HTTP: srv.Client()}
	trackURL := srv.URL + "/game-soundtracks/album/test-album/01.%20Opening.mp3"

	direct, title, err := c.resolveTrack(context.Background(), trackURL, "opus")
	if err != nil {
		t.Fatal(err)
	}
	if direct != "https://nu.vgmtreasurechest.com/soundtracks/test/01-opening.mp3" {
		t.Errorf("opus preferred direct = %q", direct)
	}
	if title != "Opening Theme" {
		t.Errorf("title = %q", title)
	}

	direct, title, err = c.resolveTrack(context.Background(), trackURL, "flac")
	if err != nil {
		t.Fatal(err)
	}
	if direct != "https://nu.vgmtreasurechest.com/soundtracks/test/01-opening.flac" {
		t.Errorf("flac preferred direct = %q", direct)
	}
	if title != "Opening Theme" {
		t.Errorf("flac resolve title = %q", title)
	}
}

func TestFetchKHInsiderAlbumInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, albumHTMLFixture)
	}))
	defer srv.Close()

	c := &khClient{HTTP: srv.Client()}
	info, err := c.fetchAlbumInfo(context.Background(), srv.URL+"/game-soundtracks/album/test-album")
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Test Game Soundtrack" {
		t.Errorf("title = %q", info.Title)
	}
	if info.Artist != "Test Publisher" {
		t.Errorf("artist = %q, want Test Publisher", info.Artist)
	}
	if info.AlbumArtist != "Test Publisher" {
		t.Errorf("album artist = %q, want Test Publisher", info.AlbumArtist)
	}
	if info.Year != "2024" || info.Genre != "Soundtrack" || info.TotalTracks != 3 {
		t.Errorf("metadata = %+v", info)
	}
	if !strings.Contains(info.CoverURL, "00%20Front.jpg") {
		t.Errorf("cover = %q", info.CoverURL)
	}
}

func TestFetchKHInsiderTrackInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, trackHTMLFixture)
	}))
	defer srv.Close()

	c := &khClient{HTTP: srv.Client()}
	info, err := c.fetchTrackInfo(context.Background(), srv.URL+"/game-soundtracks/album/test-album/1-02.%20Battle.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if info.Title != "Opening Theme" {
		t.Errorf("title = %q", info.Title)
	}
	if info.DiscNumber != "1" || info.TrackNumber != "02" {
		t.Errorf("disc/track = %q/%q", info.DiscNumber, info.TrackNumber)
	}
}

func TestKHInsiderAlbumURL(t *testing.T) {
	track := "https://downloads.khinsider.com/game-soundtracks/album/foo/01.%20Bar.mp3"
	want := "https://downloads.khinsider.com/game-soundtracks/album/foo"
	if got := KHInsiderAlbumURL(track); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestPickDirectURL(t *testing.T) {
	audio := "https://nu.vgmtreasurechest.com/a.flac"
	flac := "https://nu.vgmtreasurechest.com/b.flac"
	mp3 := "https://nu.vgmtreasurechest.com/c.mp3"

	if got := pickDirectURL("flac", audio, flac, mp3); got != flac {
		t.Errorf("flac preference = %q", got)
	}
	if got := pickDirectURL("opus", audio, flac, mp3); got != mp3 {
		t.Errorf("opus preference = %q", got)
	}
}

func TestKHTitleFromSegment(t *testing.T) {
	if got := khTitleFromSegment("01.%20Opening.mp3"); got != "01. Opening" {
		t.Errorf("got %q", got)
	}
}

func TestSanitizeKHOutputTitle(t *testing.T) {
	if got := sanitizeKHOutputTitle(`Track: "Boss"/Fight`); got != `Track_ _Boss__Fight` {
		t.Errorf("got %q", got)
	}
}

const infoTxtFixture = `Name: Kirby Star Allies
Year: 2018
Platforms: Switch
Published by: HAL Laboratory

Music by
Tadashi Ikegami (HAL Laboratory)
Hirokazu Ando (HAL Laboratory)

Disc 1
1-01. Kirby Star Allies Main Theme
`

func TestParseInfoTxt(t *testing.T) {
	title, year, publisher, platforms, composers := parseInfoTxt(infoTxtFixture)
	if title != "Kirby Star Allies" {
		t.Errorf("title = %q", title)
	}
	if year != "2018" || publisher != "HAL Laboratory" || platforms != "Switch" {
		t.Errorf("fields = %q %q %q", year, publisher, platforms)
	}
	if len(composers) != 2 {
		t.Fatalf("composers = %v", composers)
	}
	if got := artistFromComposers(composers); got != "Various Artists" {
		t.Errorf("artist = %q", got)
	}
}

func TestFetchKHInsiderAlbumInfoFromInfoTxt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "khinsider.info.txt") {
			fmt.Fprint(w, infoTxtFixture)
			return
		}
		body := strings.Replace(albumHTMLFixture,
			`</body></html>`,
			`<a href="/game-soundtracks/album/test-album/khinsider.info.txt">info</a></body></html>`, 1)
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	c := &khClient{HTTP: srv.Client()}
	info, err := c.fetchAlbumInfo(context.Background(), srv.URL+"/game-soundtracks/album/test-album")
	if err != nil {
		t.Fatal(err)
	}
	if info.Artist != "Various Artists" {
		t.Errorf("artist = %q, want Various Artists", info.Artist)
	}
	if info.AlbumArtist != "HAL Laboratory" {
		t.Errorf("album artist = %q, want HAL Laboratory", info.AlbumArtist)
	}
	if info.Title != "Kirby Star Allies" {
		t.Errorf("title = %q", info.Title)
	}
}
