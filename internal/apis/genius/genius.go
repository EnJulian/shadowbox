// Package genius is a minimal client for the Genius API. It searches for a song
// and scrapes the public lyrics page, since the API itself does not return
// lyrics text.
package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Client talks to the Genius API and lyrics pages.
type Client struct {
	Token   string
	HTTP    *http.Client
	APIURL  string // overridable for testing
	PageURL string // optional base override for lyric page fetches (testing)
}

// New returns a Genius client for the given access token.
func New(token string) *Client {
	return &Client{
		Token:  token,
		HTTP:   &http.Client{Timeout: 15 * time.Second},
		APIURL: "https://api.genius.com",
	}
}

// Configured reports whether an access token is present.
func (c *Client) Configured() bool { return c.Token != "" }

type searchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				URL           string `json:"url"`
				Title         string `json:"title"`
				PrimaryArtist struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

// Lyrics searches Genius for the track and returns cleaned lyrics text, or ""
// when none are found.
func (c *Client) Lyrics(ctx context.Context, title, artist string) (string, error) {
	if !c.Configured() {
		return "", nil
	}
	pageURL, err := c.searchSongURL(ctx, title, artist)
	if err != nil {
		return "", err
	}
	if pageURL == "" {
		return "", nil
	}
	return c.scrapeLyrics(ctx, pageURL)
}

func (c *Client) searchSongURL(ctx context.Context, title, artist string) (string, error) {
	q := title
	if artist != "" {
		q = title + " " + artist
	}
	u := fmt.Sprintf("%s/search?q=%s", c.APIURL, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("genius search failed: %s", resp.Status)
	}
	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", err
	}
	if len(sr.Response.Hits) == 0 {
		return "", nil
	}
	return sr.Response.Hits[0].Result.URL, nil
}

func (c *Client) scrapeLyrics(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "shadowbox/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("genius page fetch failed: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return extractLyrics(string(body)), nil
}

// extractLyrics parses Genius HTML, pulling text from any element marked with
// data-lyrics-container="true", then cleans up the result.
func extractLyrics(htmlBody string) string {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return ""
	}
	var sb strings.Builder
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inContainer bool) {
		if n.Type == html.ElementNode {
			if !inContainer {
				for _, a := range n.Attr {
					if a.Key == "data-lyrics-container" && a.Val == "true" {
						inContainer = true
						break
					}
				}
			}
			if n.Data == "br" && inContainer {
				sb.WriteByte('\n')
			}
		}
		if n.Type == html.TextNode && inContainer {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inContainer)
		}
	}
	walk(doc, false)
	return cleanLyrics(sb.String())
}

var (
	embedRe       = regexp.MustCompile(`\d+Embed\s*$`)
	multiNewline  = regexp.MustCompile(`\n\s*\n\s*\n+`)
	alsoLikeRe    = regexp.MustCompile(`(?i)You might also like`)
)

func cleanLyrics(s string) string {
	s = strings.TrimSpace(s)
	s = alsoLikeRe.ReplaceAllString(s, "")
	s = embedRe.ReplaceAllString(s, "")
	s = multiNewline.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
