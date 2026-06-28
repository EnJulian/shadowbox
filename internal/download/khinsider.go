package download

import (
	"context"
	"fmt"
	htmlpkg "html"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EnJulian/shadowbox/internal/organize"
	"golang.org/x/net/html"
)

const khUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// khScrapeDelay is the pause between successive KHInsider track page fetches.
const khScrapeDelay = 500 * time.Millisecond

var khTrackPageSuffix = map[string]bool{".mp3": true, ".flac": true}

var (
	khYearRe         = regexp.MustCompile(`(?i)Year:\s*<b>([^<]+)</b>`)
	khNumFilesRe     = regexp.MustCompile(`(?i)Number of Files:\s*<b>(\d+)</b>`)
	khSongNameRe     = regexp.MustCompile(`(?i)Song name:\s*<b>([^<]+)</b>`)
	khDiscTrackRe    = regexp.MustCompile(`^(\d+)-(\d+)`)
	khLeadingTrackRe = regexp.MustCompile(`^(\d+)[.\s_-]`)
)

const khMaxCoverBytes = 10 << 20 // 10 MiB

// KHInsiderAlbumInfo holds album-level metadata scraped from a KHInsider page.
type KHInsiderAlbumInfo struct {
	AlbumURL    string
	Title       string
	Artist      string
	AlbumArtist string
	Composer    string
	Composers   []string
	Publisher   string
	Platforms   string
	Year        string
	Genre       string
	TotalTracks int
	TotalDiscs  int
	CoverURL    string
	Tracks      []khTrack
}

// KHInsiderTrackInfo holds track-level metadata scraped from a KHInsider track page.
type KHInsiderTrackInfo struct {
	Title       string
	TrackNumber string
	DiscNumber  string
}

type khTrack struct {
	Index   int
	Title   string
	PageURL string
}

type khClient struct {
	HTTP *http.Client
}

func defaultKHClient() *khClient {
	return &khClient{HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *khClient) fetch(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", khUserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("KHInsider fetch failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *khClient) fetchBytes(ctx context.Context, pageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", khUserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("KHInsider fetch failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// KHInsiderAlbumURL returns the album page URL for a KHInsider album or track URL.
func KHInsiderAlbumURL(rawURL string) string {
	if IsKHInsiderPlaylist(rawURL) {
		return rawURL
	}
	if !IsKHInsiderTrack(rawURL) {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	segs := khPathSegments(rawURL)
	if len(segs) < 3 {
		return ""
	}
	u.Path = "/" + strings.Join(segs[:3], "/")
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// FetchKHInsiderAlbumInfo scrapes album metadata and track list from a KHInsider album URL.
func FetchKHInsiderAlbumInfo(ctx context.Context, albumURL string) (*KHInsiderAlbumInfo, error) {
	return defaultKHClient().fetchAlbumInfo(ctx, albumURL)
}

func (c *khClient) fetchAlbumInfo(ctx context.Context, albumURL string) (*KHInsiderAlbumInfo, error) {
	body, err := c.fetchBytes(ctx, albumURL)
	if err != nil {
		return nil, err
	}
	htmlBody := string(body)
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("parse album page: %w", err)
	}

	info := &KHInsiderAlbumInfo{
		AlbumURL: albumURL,
		CoverURL: findAlbumCoverURL(doc),
	}
	c.enrichAlbumFromInfoTxt(ctx, doc, info, albumURL)
	applyHTMLAlbumFields(info, doc, htmlBody)
	if m := khNumFilesRe.FindStringSubmatch(htmlBody); len(m) == 2 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			info.TotalTracks = n
		}
	}

	base, err := url.Parse(albumURL)
	if err != nil {
		return nil, fmt.Errorf("invalid album URL: %w", err)
	}
	info.Tracks, err = c.parseAlbumTracks(doc, base)
	if err != nil {
		return nil, err
	}
	if info.TotalTracks == 0 {
		info.TotalTracks = len(info.Tracks)
	}
	info.finalizeMetadata()
	return info, nil
}

// FetchKHInsiderTrackInfo scrapes track title and disc/track numbers from a track page.
func FetchKHInsiderTrackInfo(ctx context.Context, trackURL string) (*KHInsiderTrackInfo, error) {
	return defaultKHClient().fetchTrackInfo(ctx, trackURL)
}

func (c *khClient) fetchTrackInfo(ctx context.Context, trackURL string) (*KHInsiderTrackInfo, error) {
	body, err := c.fetchBytes(ctx, trackURL)
	if err != nil {
		return nil, err
	}
	htmlBody := string(body)
	info := &KHInsiderTrackInfo{}
	if m := khSongNameRe.FindStringSubmatch(htmlBody); len(m) == 2 {
		info.Title = strings.TrimSpace(htmlUnescape(m[1]))
	}
	segs := khPathSegments(trackURL)
	if len(segs) > 0 {
		info.DiscNumber, info.TrackNumber = parseDiscTrackNumbers(segs[len(segs)-1])
	}
	if info.Title == "" && len(segs) > 0 {
		info.Title = khTitleFromSegment(segs[len(segs)-1])
	}
	return info, nil
}

// KHInsiderTrackFromAlbum builds track metadata from scraped album data without
// an extra HTTP request.
func KHInsiderTrackFromAlbum(pageURL, title string, index int) *KHInsiderTrackInfo {
	info := &KHInsiderTrackInfo{Title: title}
	if index > 0 {
		info.TrackNumber = strconv.Itoa(index)
	}
	segs := khPathSegments(pageURL)
	if len(segs) > 0 {
		disc, trackNum := parseDiscTrackNumbers(segs[len(segs)-1])
		if disc != "" {
			info.DiscNumber = disc
		}
		if trackNum != "" {
			info.TrackNumber = trackNum
		}
	}
	return info
}

// FetchKHCover downloads the album cover image bytes from a mirror URL.
func FetchKHCover(ctx context.Context, coverURL string) ([]byte, string, error) {
	return defaultKHClient().downloadCover(ctx, coverURL)
}

func (c *khClient) downloadCover(ctx context.Context, coverURL string) ([]byte, string, error) {
	u, err := url.Parse(coverURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid cover URL: %w", err)
	}
	if strings.ToLower(u.Scheme) != "https" {
		return nil, "", fmt.Errorf("cover URL must use HTTPS")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", khUserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("cover download failed: %s", resp.Status)
	}

	mime := resp.Header.Get("Content-Type")
	if mime == "" || !strings.HasPrefix(strings.ToLower(mime), "image/") {
		return nil, "", fmt.Errorf("cover response is not an image")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, khMaxCoverBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > khMaxCoverBytes {
		return nil, "", fmt.Errorf("cover image exceeds size limit")
	}
	return data, mime, nil
}

func extractKHField(re *regexp.Regexp, htmlBody string) string {
	if m := re.FindStringSubmatch(htmlBody); len(m) == 2 {
		return strings.TrimSpace(htmlUnescape(m[1]))
	}
	return ""
}

func htmlUnescape(s string) string {
	return htmlpkg.UnescapeString(s)
}

func findAlbumCoverURL(doc *html.Node) string {
	var fullLinks []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "albumImage") {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "a" {
					if href := attrVal(child, "href"); isKHCoverLink(href) {
						fullLinks = append(fullLinks, href)
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	prefer := []string{"front", "cover", "display", "box front"}
	for _, key := range prefer {
		for _, link := range fullLinks {
			if strings.Contains(strings.ToLower(link), key) {
				return link
			}
		}
	}
	if len(fullLinks) > 0 {
		return fullLinks[0]
	}
	return ""
}

func isKHCoverLink(href string) bool {
	if href == "" {
		return false
	}
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	if !strings.HasSuffix(strings.ToLower(u.Hostname()), "vgmtreasurechest.com") {
		return false
	}
	if strings.Contains(strings.ToLower(u.Path), "/thumbs/") {
		return false
	}
	ext := strings.ToLower(path.Ext(u.Path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

func hasClass(n *html.Node, want string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" && a.Val == want {
			return true
		}
	}
	return false
}

func parseDiscTrackNumbers(segment string) (disc, track string) {
	decoded, err := url.PathUnescape(segment)
	if err != nil {
		decoded = segment
	}
	decoded = strings.TrimSuffix(decoded, path.Ext(decoded))
	if m := khDiscTrackRe.FindStringSubmatch(decoded); len(m) == 3 {
		return m[1], m[2]
	}
	if m := khLeadingTrackRe.FindStringSubmatch(decoded); len(m) == 2 {
		return "1", m[1]
	}
	return "", ""
}

func (c *khClient) parseAlbumTracks(doc *html.Node, base *url.URL) ([]khTrack, error) {
	seen := make(map[string]bool)
	var tracks []khTrack
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := attrVal(n, "href")
			if href != "" {
				if track, ok := parseKHTrackLink(href, n, base); ok && !seen[track.PageURL] {
					seen[track.PageURL] = true
					track.Index = len(tracks) + 1
					segs := khPathSegments(track.PageURL)
					if len(segs) > 0 {
						_, trackNum := parseDiscTrackNumbers(segs[len(segs)-1])
						if trackNum != "" {
							track.Title = strings.TrimSpace(strings.TrimPrefix(track.Title, trackNum+". "))
							track.Title = strings.TrimSpace(strings.TrimPrefix(track.Title, trackNum+" - "))
						}
					}
					tracks = append(tracks, track)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return tracks, nil
}

func scrapeKHInsiderAlbum(ctx context.Context, albumURL string) ([]khTrack, error) {
	info, err := defaultKHClient().fetchAlbumInfo(ctx, albumURL)
	if err != nil {
		return nil, err
	}
	if len(info.Tracks) == 0 {
		return nil, fmt.Errorf("no tracks found on KHInsider album page")
	}
	return info.Tracks, nil
}

func parseKHTrackLink(href string, n *html.Node, base *url.URL) (khTrack, bool) {
	ref, err := url.Parse(href)
	if err != nil {
		return khTrack{}, false
	}
	abs := base.ResolveReference(ref)
	segs := khPathSegments(abs.String())
	if len(segs) <= 3 || segs[0] != "game-soundtracks" || segs[1] != "album" {
		return khTrack{}, false
	}
	ext := strings.ToLower(path.Ext(segs[len(segs)-1]))
	if !khTrackPageSuffix[ext] {
		return khTrack{}, false
	}
	title := strings.TrimSpace(textContent(n))
	if title == "" {
		title = khTitleFromSegment(segs[len(segs)-1])
	}
	return khTrack{
		Title:   title,
		PageURL: abs.String(),
	}, true
}

func resolveKHInsiderTrack(ctx context.Context, trackURL, preferredFormat string) (directURL, title string, err error) {
	return defaultKHClient().resolveTrack(ctx, trackURL, preferredFormat)
}

func (c *khClient) resolveTrack(ctx context.Context, trackURL, preferredFormat string) (directURL, title string, err error) {
	htmlBody, err := c.fetch(ctx, trackURL)
	if err != nil {
		return "", "", err
	}
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return "", "", fmt.Errorf("parse track page: %w", err)
	}

	audioSrc := findAudioSrc(doc)
	flacLink, mp3Link := findMirrorLinks(doc)

	directURL = pickDirectURL(preferredFormat, audioSrc, flacLink, mp3Link)
	if directURL == "" {
		return "", "", fmt.Errorf("no direct download link found on KHInsider track page")
	}

	if m := khSongNameRe.FindStringSubmatch(htmlBody); len(m) == 2 {
		title = strings.TrimSpace(htmlUnescape(m[1]))
	}
	if title == "" {
		segs := khPathSegments(trackURL)
		if len(segs) > 0 {
			title = khTitleFromSegment(segs[len(segs)-1])
		} else {
			title = "Unknown Track"
		}
	}
	return directURL, title, nil
}

func pickDirectURL(preferredFormat, audioSrc, flacLink, mp3Link string) string {
	wantFlac := strings.EqualFold(preferredFormat, "flac")
	if wantFlac {
		if flacLink != "" {
			return flacLink
		}
		if audioSrc != "" && strings.HasSuffix(strings.ToLower(audioSrc), ".flac") {
			return audioSrc
		}
		if mp3Link != "" {
			return mp3Link
		}
		return audioSrc
	}
	if mp3Link != "" {
		return mp3Link
	}
	if audioSrc != "" && strings.HasSuffix(strings.ToLower(audioSrc), ".mp3") {
		return audioSrc
	}
	if flacLink != "" {
		return flacLink
	}
	return audioSrc
}

func findAudioSrc(doc *html.Node) string {
	var found string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "audio" {
			if id := attrVal(n, "id"); id == "audio" || found == "" {
				if src := attrVal(n, "src"); src != "" {
					found = src
					if id == "audio" {
						return
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return found
}

func findMirrorLinks(doc *html.Node) (flac, mp3 string) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := attrVal(n, "href")
			if isMirrorMediaLink(href) {
				lower := strings.ToLower(href)
				if strings.HasSuffix(lower, ".flac") && flac == "" {
					flac = href
				}
				if strings.HasSuffix(lower, ".mp3") && mp3 == "" {
					mp3 = href
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return flac, mp3
}

func isMirrorMediaLink(href string) bool {
	if href == "" {
		return false
	}
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if !strings.HasSuffix(host, "vgmtreasurechest.com") {
		return false
	}
	ext := strings.ToLower(path.Ext(u.Path))
	return ext == ".mp3" || ext == ".flac"
}

func khTitleFromSegment(segment string) string {
	decoded, err := url.PathUnescape(segment)
	if err != nil {
		decoded = segment
	}
	ext := path.Ext(decoded)
	title := strings.TrimSuffix(decoded, ext)
	title = strings.TrimSpace(title)
	if title == "" {
		return "Unknown Track"
	}
	return title
}

func sanitizeKHOutputTitle(title string) string {
	return organize.SanitizeFilename(title)
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		sb.WriteString(textContent(child))
	}
	return sb.String()
}
