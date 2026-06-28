package download

import (
	"context"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

var (
	khPublishedRe = regexp.MustCompile(`(?is)Published by:\s*(?:<a[^>]*>([^<]+)</a>|<b>([^<]+)</b>|([^<\n]+))`)
	khAlbumTypeRe = regexp.MustCompile(`(?is)Album type:\s*<b>(?:\s*<a[^>]*>([^<]+)</a>|([^<]+?))\s*</b>`)
	khPlatformsRe = regexp.MustCompile(`(?is)Platforms:\s*(?:<a[^>]*>([^<]+)</a>|([^<,\n]+))`)
	khComposerLineRe = regexp.MustCompile(`(?is)Composer:\s*([^<\n]+)`)
)

func extractKHFieldGroups(re *regexp.Regexp, htmlBody string) string {
	m := re.FindStringSubmatch(htmlBody)
	if m == nil {
		return ""
	}
	for i := 1; i < len(m); i++ {
		if s := strings.TrimSpace(htmlUnescape(m[i])); s != "" {
			return s
		}
	}
	return ""
}

func findInfoTxtURL(doc *html.Node) string {
	var found string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "a" {
			if href := attrVal(n, "href"); strings.Contains(strings.ToLower(href), "khinsider.info.txt") {
				found = href
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return found
}

func parseInfoTxt(body string) (title, year, publisher, platforms string, composers []string) {
	lines := strings.Split(body, "\n")
	inMusicBy := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "Name:"):
			title = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		case strings.HasPrefix(line, "Year:"):
			year = strings.TrimSpace(strings.TrimPrefix(line, "Year:"))
		case strings.HasPrefix(line, "Published by:"):
			publisher = strings.TrimSpace(strings.TrimPrefix(line, "Published by:"))
		case strings.HasPrefix(line, "Platforms:"):
			platforms = strings.TrimSpace(strings.TrimPrefix(line, "Platforms:"))
		case line == "Music by":
			inMusicBy = true
		case strings.HasPrefix(line, "Disc "):
			inMusicBy = false
		case strings.HasPrefix(line, "Description:"):
			inMusicBy = false
		case inMusicBy:
			if name := composerNameFromCredit(line); name != "" {
				composers = appendUnique(composers, name)
			}
		}
	}
	return title, year, publisher, platforms, composers
}

func composerNameFromCredit(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	if i := strings.Index(line, "("); i > 0 {
		line = strings.TrimSpace(line[:i])
	}
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	if line == "" || strings.HasPrefix(strings.ToLower(line), "vocals:") {
		return ""
	}
	return line
}

func appendUnique(list []string, val string) []string {
	for _, existing := range list {
		if strings.EqualFold(existing, val) {
			return list
		}
	}
	return append(list, val)
}

func artistFromComposers(composers []string) string {
	switch len(composers) {
	case 0:
		return ""
	case 1:
		return composers[0]
	default:
		return "Various Artists"
	}
}

func mergeInfoTxtFields(info *KHInsiderAlbumInfo, title, year, publisher, platforms string, composers []string) {
	if title != "" {
		info.Title = title
	}
	if year != "" {
		info.Year = year
	}
	if publisher != "" {
		info.Publisher = publisher
	}
	if platforms != "" {
		info.Platforms = platforms
	}
	if len(composers) > 0 {
		info.Composers = composers
		info.Artist = artistFromComposers(composers)
	}
}

func applyHTMLAlbumFields(info *KHInsiderAlbumInfo, doc *html.Node, htmlBody string) {
	if info.Title == "" {
		info.Title = strings.TrimSpace(htmlUnescape(firstAlbumTitle(doc)))
	}
	if info.Year == "" {
		info.Year = extractKHField(khYearRe, htmlBody)
	}
	if info.Publisher == "" {
		info.Publisher = extractKHFieldGroups(khPublishedRe, htmlBody)
	}
	if info.Genre == "" {
		info.Genre = extractKHFieldGroups(khAlbumTypeRe, htmlBody)
	}
	if info.Platforms == "" {
		info.Platforms = extractKHFieldGroups(khPlatformsRe, htmlBody)
	}
	if info.Composer == "" {
		if c := extractKHFieldGroups(khComposerLineRe, htmlBody); c != "" {
			info.Composer = c
			if name := composerNameFromCredit(c); name != "" && info.Artist == "" {
				info.Artist = name
			}
		}
	}
}

func firstAlbumTitle(doc *html.Node) string {
	var found string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "h2" {
			title := strings.TrimSpace(textContent(n))
			lower := strings.ToLower(title)
			if lower == "description" || lower == "comments" || lower == "ooops!" {
				return
			}
			found = title
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
		 walk(child)
		}
	}
	walk(doc)
	return found
}

func (info *KHInsiderAlbumInfo) finalizeMetadata() {
	if info.Genre == "" && info.Platforms != "" {
		info.Genre = "Video Game"
	}
	if info.Artist == "" && info.Composer != "" {
		if name := composerNameFromCredit(info.Composer); name != "" {
			info.Artist = name
		}
	}
	if info.Artist == "" && len(info.Composers) > 0 {
		info.Artist = artistFromComposers(info.Composers)
	}
	if info.Artist == "" && info.Publisher != "" {
		info.Artist = info.Publisher
	}
	if info.AlbumArtist == "" {
		info.AlbumArtist = info.Publisher
	}
	if info.AlbumArtist == "" {
		info.AlbumArtist = info.Artist
	}
	if info.TotalDiscs == 0 {
		info.TotalDiscs = maxDiscFromTracks(info.Tracks)
	}
}

func maxDiscFromTracks(tracks []khTrack) int {
	maxDisc := 0
	for _, t := range tracks {
		segs := khPathSegments(t.PageURL)
		if len(segs) == 0 {
			continue
		}
		disc, _ := parseDiscTrackNumbers(segs[len(segs)-1])
		if disc == "" {
			continue
		}
		if n := atoiSafe(disc); n > maxDisc {
			maxDisc = n
		}
	}
	return maxDisc
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func (c *khClient) enrichAlbumFromInfoTxt(ctx context.Context, doc *html.Node, info *KHInsiderAlbumInfo, albumURL string) {
	href := findInfoTxtURL(doc)
	if href == "" {
		return
	}
	base, err := url.Parse(albumURL)
	if err != nil {
		return
	}
	ref, err := url.Parse(href)
	if err != nil {
		return
	}
	body, err := c.fetchBytes(ctx, base.ResolveReference(ref).String())
	if err != nil {
		return
	}
	title, year, publisher, platforms, composers := parseInfoTxt(string(body))
	mergeInfoTxtFields(info, title, year, publisher, platforms, composers)
}
