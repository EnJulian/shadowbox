// Package musicbrainz is a minimal client for the MusicBrainz API and Cover Art
// Archive, used as Shadowbox's primary metadata source.
package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EnJulian/shadowbox/internal/apis"
)

const (
	defaultUserAgent = "shadowbox/1.0 (https://github.com/EnJulian/shadowbox)"
	minRequestGap    = time.Second
)

// Client queries the MusicBrainz Web Service and Cover Art Archive.
type Client struct {
	HTTP      *http.Client
	BaseURL   string // overridable for testing (MusicBrainz ws/2)
	CoverBase string // overridable for testing (Cover Art Archive)
	UserAgent string

	mu      sync.Mutex
	lastReq time.Time
	cache   map[string]*cacheEntry
}

type cacheEntry struct {
	meta      *apis.TrackMeta
	releaseID string
}

// New returns a MusicBrainz client with default endpoints.
func New() *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: 15 * time.Second},
		BaseURL:   "https://musicbrainz.org/ws/2",
		CoverBase: "https://coverartarchive.org",
		UserAgent: defaultUserAgent,
		cache:     make(map[string]*cacheEntry),
	}
}

type searchResponse struct {
	Recordings []recording `json:"recordings"`
}

type recording struct {
	ID               string         `json:"id"`
	Score            int            `json:"score"`
	Title            string         `json:"title"`
	Length           int            `json:"length"`
	Disambiguation   string         `json:"disambiguation"`
	Video            *bool          `json:"video"`
	FirstReleaseDate string         `json:"first-release-date"`
	ArtistCredit     []artistCredit `json:"artist-credit"`
	Releases         []release      `json:"releases"`
	Tags             []tag          `json:"tags"`
}

type artistCredit struct {
	Name       string `json:"name"`
	JoinPhrase string `json:"joinphrase"`
	Artist     struct {
		Name string `json:"name"`
	} `json:"artist"`
}

type release struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Status       string       `json:"status"`
	Date         string       `json:"date"`
	TrackCount   int          `json:"track-count"`
	ReleaseGroup releaseGroup `json:"release-group"`
	Media        []medium     `json:"media"`
	Tags         []tag        `json:"tags"`
}

type releaseGroup struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	PrimaryType    string   `json:"primary-type"`
	SecondaryTypes []string `json:"secondary-types"`
	Tags           []tag    `json:"tags"`
}

type medium struct {
	Position   int     `json:"position"`
	Format     string  `json:"format"`
	TrackCount int     `json:"track-count"`
	Tracks     []track `json:"track"`
}

type track struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Title  string `json:"title"`
	Length int    `json:"length"`
}

type tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type coverResponse struct {
	Images []coverImage `json:"images"`
}

type coverImage struct {
	Front      bool              `json:"front"`
	Thumbnails map[string]string `json:"thumbnails"`
	Image      string            `json:"image"`
}

// RecordingCandidate is one MusicBrainz recording match for user selection.
type RecordingCandidate struct {
	ID     string
	Title  string
	Artist string
	Album  string
	Date   string
	Length string

	rec recording
}

// SearchCandidates returns recording matches without auto-picking one.
func (c *Client) SearchCandidates(ctx context.Context, title, artist string) ([]RecordingCandidate, error) {
	var res *searchResponse
	var err error
	for _, q := range buildQueries(title, artist) {
		res, err = c.searchRecordings(ctx, q)
		if err != nil {
			return nil, err
		}
		if len(res.Recordings) > 0 {
			break
		}
	}
	if res == nil || len(res.Recordings) == 0 {
		return nil, nil
	}

	out := make([]RecordingCandidate, 0, len(res.Recordings))
	for _, rec := range res.Recordings {
		rel, _, _ := pickRelease(rec, title)
		album := ""
		date := rec.FirstReleaseDate
		if rel != nil {
			album = albumTitle(rel)
			date = releaseDate(rel, rec)
		}
		out = append(out, RecordingCandidate{
			ID:     rec.ID,
			Title:  rec.Title,
			Artist: joinArtistCredit(rec.ArtistCredit),
			Album:  album,
			Date:   yearFromDate(date),
			Length: formatLength(rec.Length),
			rec:    rec,
		})
	}
	return out, nil
}

// MetaFromCandidate builds normalised metadata for a chosen recording candidate.
func (c *Client) MetaFromCandidate(ctx context.Context, cand RecordingCandidate, titleHint string) (*apis.TrackMeta, error) {
	return c.metaFromRecording(ctx, cand.rec, titleHint)
}

// Search looks up a track by title and optional artist, returning normalised
// metadata. It returns (nil, nil) when nothing is found.
func (c *Client) Search(ctx context.Context, title, artist string) (*apis.TrackMeta, error) {
	key := cacheKey(title, artist)
	if e := c.cached(key); e != nil {
		return e.meta, nil
	}

	candidates, err := c.SearchCandidates(ctx, title, artist)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	rec := pickRecording(recordingsFromCandidates(candidates))
	meta, err := c.metaFromRecording(ctx, rec, title)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, nil
	}

	releaseID := ""
	if rel, _, _ := pickRelease(rec, title); rel != nil {
		releaseID = rel.ID
	}
	c.store(key, &cacheEntry{meta: meta, releaseID: releaseID})
	return meta, nil
}

func recordingsFromCandidates(candidates []RecordingCandidate) []recording {
	out := make([]recording, len(candidates))
	for i, c := range candidates {
		out[i] = c.rec
	}
	return out
}

func (c *Client) metaFromRecording(ctx context.Context, rec recording, title string) (*apis.TrackMeta, error) {
	rel, trk, med := pickRelease(rec, title)
	if rel == nil {
		return nil, nil
	}

	meta := &apis.TrackMeta{
		Title:       rec.Title,
		Artist:      joinArtistCredit(rec.ArtistCredit),
		Album:       albumTitle(rel),
		ReleaseDate: releaseDate(rel, rec),
		Performer:   joinArtistCredit(rec.ArtistCredit),
		Genre:       topTag(rec.Tags, rel.Tags, rel.ReleaseGroup.Tags),
	}

	if trk != nil {
		meta.TrackNumber = trk.Number
	}
	if med != nil {
		if med.Position > 0 {
			meta.DiscNumber = strconv.Itoa(med.Position)
		}
		if med.TrackCount > 0 {
			meta.TotalTracks = strconv.Itoa(med.TrackCount)
		}
	}
	if rel.TrackCount > 0 && meta.TotalTracks == "" {
		meta.TotalTracks = strconv.Itoa(rel.TrackCount)
	}
	if len(rel.Media) > 0 {
		meta.TotalDiscs = strconv.Itoa(len(rel.Media))
	}

	if rel.ID != "" {
		if u, err := c.CoverURL(ctx, rel.ID); err == nil {
			meta.CoverURL = u
		}
	}

	return meta, nil
}

// CoverURL returns a front cover image URL for the given release MBID.
func (c *Client) CoverURL(ctx context.Context, releaseID string) (string, error) {
	if strings.TrimSpace(releaseID) == "" {
		return "", nil
	}
	u := fmt.Sprintf("%s/release/%s", strings.TrimRight(c.CoverBase, "/"), url.PathEscape(releaseID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", c.userAgent())
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cover art archive request failed: %s", resp.Status)
	}

	var cr coverResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", err
	}
	return pickCoverURL(cr.Images), nil
}

func (c *Client) searchRecordings(ctx context.Context, query string) (*searchResponse, error) {
	params := url.Values{
		"query":  {query},
		"fmt":    {"json"},
		"limit":  {"5"},
		"dismax": {"true"},
	}
	u := c.BaseURL + "/recording?" + params.Encode()

	var res searchResponse
	if err := c.getJSON(ctx, u, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) getJSON(ctx context.Context, u string, out any) error {
	if err := c.throttle(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent())
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("musicbrainz request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) throttle(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.lastReq.IsZero() {
		wait := minRequestGap - time.Since(c.lastReq)
		if wait > 0 {
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	c.lastReq = time.Now()
	return nil
}

func (c *Client) userAgent() string {
	if c.UserAgent != "" {
		return c.UserAgent
	}
	return defaultUserAgent
}

func (c *Client) cached(key string) *cacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache[key]
}

func (c *Client) store(key string, e *cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = e
}

func cacheKey(title, artist string) string {
	return strings.ToLower(strings.TrimSpace(title)) + "|" + strings.ToLower(strings.TrimSpace(artist))
}

func buildQueries(title, artist string) []string {
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)
	titleQ := luceneQuote(title)

	var queries []string
	if artist != "" {
		artistQ := luceneQuote(artist)
		queries = append(queries,
			fmt.Sprintf(`recording:%s AND artist:%s AND status:official AND -video:true`, titleQ, artistQ),
			fmt.Sprintf(`recording:%s AND artist:%s`, titleQ, artistQ),
		)
	}
	queries = append(queries, fmt.Sprintf(`recording:%s`, titleQ))
	return queries
}

func luceneQuote(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func pickRecording(recordings []recording) recording {
	best := recordings[0]
	bestScore := scoreRecording(best)
	for _, rec := range recordings[1:] {
		if s := scoreRecording(rec); s > bestScore {
			best = rec
			bestScore = s
		}
	}
	return best
}

func scoreRecording(rec recording) int {
	score := rec.Score * 10
	if rec.Disambiguation == "" {
		score += 50
	} else if strings.Contains(strings.ToLower(rec.Disambiguation), "live") {
		score -= 30
	}
	if rec.Video != nil && *rec.Video {
		score -= 40
	}
	return score
}

func pickRelease(rec recording, title string) (*release, *track, *medium) {
	if len(rec.Releases) == 0 {
		return nil, nil, nil
	}
	type candidate struct {
		rel   release
		trk   *track
		med   *medium
		score int
	}
	var candidates []candidate
	titleLower := strings.ToLower(strings.TrimSpace(title))
	for _, rel := range rec.Releases {
		trk, med := findTrack(rel, rec.Title, titleLower)
		candidates = append(candidates, candidate{
			rel:   rel,
			trk:   trk,
			med:   med,
			score: scoreRelease(rel),
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	best := candidates[0]
	return &best.rel, best.trk, best.med
}

func findTrack(rel release, recTitle, queryTitle string) (*track, *medium) {
	recLower := strings.ToLower(recTitle)
	for i := range rel.Media {
		med := &rel.Media[i]
		for j := range med.Tracks {
			trk := &med.Tracks[j]
			tl := strings.ToLower(trk.Title)
			if tl == recLower || tl == queryTitle || strings.Contains(tl, queryTitle) || strings.Contains(queryTitle, tl) {
				return trk, med
			}
		}
	}
	if len(rel.Media) > 0 && len(rel.Media[0].Tracks) > 0 {
		return &rel.Media[0].Tracks[0], &rel.Media[0]
	}
	return nil, nil
}

func scoreRelease(rel release) int {
	score := 0
	switch strings.ToLower(rel.Status) {
	case "official":
		score += 100
	case "promotion":
		score += 40
	case "bootleg":
		score -= 80
	default:
		score -= 20
	}
	for _, st := range rel.ReleaseGroup.SecondaryTypes {
		switch strings.ToLower(st) {
		case "compilation":
			score -= 25
		case "live":
			score -= 35
		case "soundtrack":
			score -= 10
		}
	}
	if rel.Date != "" {
		score += 10
	}
	for _, med := range rel.Media {
		f := strings.ToLower(med.Format)
		if strings.Contains(f, "dvd") || strings.Contains(f, "vhs") || strings.Contains(f, "video") {
			score -= 30
		}
		if strings.Contains(f, "cd") || strings.Contains(f, "vinyl") || strings.Contains(f, "digital") {
			score += 5
		}
	}
	return score
}

func joinArtistCredit(credits []artistCredit) string {
	if len(credits) == 0 {
		return ""
	}
	var b strings.Builder
	for i, ac := range credits {
		if i > 0 {
			b.WriteString(ac.JoinPhrase)
		}
		name := ac.Name
		if name == "" {
			name = ac.Artist.Name
		}
		b.WriteString(name)
	}
	return b.String()
}

func albumTitle(rel *release) string {
	if rel == nil {
		return ""
	}
	if rel.ReleaseGroup.Title != "" {
		return rel.ReleaseGroup.Title
	}
	return rel.Title
}

func releaseDate(rel *release, rec recording) string {
	if rel != nil && rel.Date != "" {
		return rel.Date
	}
	return rec.FirstReleaseDate
}

func topTag(tagSets ...[]tag) string {
	counts := make(map[string]int)
	for _, tags := range tagSets {
		for _, t := range tags {
			name := strings.TrimSpace(t.Name)
			if name == "" {
				continue
			}
			if t.Count > counts[name] {
				counts[name] = t.Count
			}
		}
	}
	if len(counts) == 0 {
		return ""
	}
	type kv struct {
		name  string
		count int
	}
	var ranked []kv
	for name, count := range counts {
		ranked = append(ranked, kv{name, count})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].count == ranked[j].count {
			return ranked[i].name < ranked[j].name
		}
		return ranked[i].count > ranked[j].count
	})
	return ranked[0].name
}

func pickCoverURL(images []coverImage) string {
	var front, largest coverImage
	largestArea := -1
	for _, img := range images {
		if img.Front {
			front = img
			break
		}
		area := thumbnailArea(img.Thumbnails)
		if area > largestArea {
			largestArea = area
			largest = img
		}
	}
	target := front
	if target.Image == "" && target.Thumbnails == nil {
		target = largest
	}
	if u := bestThumbnail(target.Thumbnails); u != "" {
		return u
	}
	return target.Image
}

func bestThumbnail(thumbs map[string]string) string {
	if len(thumbs) == 0 {
		return ""
	}
	order := []string{"large", "1200", "1000", "500", "400", "250", "small"}
	for _, k := range order {
		if u := thumbs[k]; u != "" && strings.HasPrefix(strings.ToLower(u), "https://") {
			return u
		}
	}
	best := ""
	bestArea := -1
	for size, u := range thumbs {
		if !strings.HasPrefix(strings.ToLower(u), "https://") {
			continue
		}
		area := 0
		if n, err := strconv.Atoi(size); err == nil {
			area = n * n
		}
		if area > bestArea {
			bestArea = area
			best = u
		}
	}
	return best
}

func thumbnailArea(thumbs map[string]string) int {
	best := 0
	for size := range thumbs {
		if n, err := strconv.Atoi(size); err == nil && n*n > best {
			best = n * n
		}
	}
	return best
}

func formatLength(ms int) string {
	if ms <= 0 {
		return ""
	}
	total := ms / 1000
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}

func yearFromDate(date string) string {
	date = strings.TrimSpace(date)
	if len(date) >= 4 {
		return date[:4]
	}
	return date
}
