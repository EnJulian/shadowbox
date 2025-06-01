# YouTube Download Issues - Troubleshooting Guide

## Common Error: "Sign in to confirm you're not a bot"

If you're seeing this error when trying to download from YouTube:

```
ERROR: [youtube] pFZHoIdcvhY: Sign in to confirm you're not a bot. This helps protect our community. Learn more
```

This is a common issue caused by YouTube's anti-bot detection measures. Here are several solutions:

## Quick Solutions

### 1. Update yt-dlp (Recommended)

The most effective solution is to update yt-dlp to the latest version:

**Using Shadowbox Settings Menu:**
1. Run Shadowbox
2. Go to Settings (option 4)
3. Select "Update yt-dlp" (option 5)
4. Follow the prompts

**Using Command Line:**

**Windows:**
```bash
# Run the update script
setup\update_ytdlp.bat

# Or manually:
pip install --upgrade yt-dlp
```

**Linux/macOS:**
```bash
# Run the update script
./setup/update_ytdlp.sh

# Or manually:
pip3 install --upgrade yt-dlp
```

**Python:**
```bash
# Run the Python updater
python3 setup/update_ytdlp.py
```

### 2. Try Different Search Terms

Sometimes the issue is specific to certain videos or search terms:
- Try searching for the same song with different keywords
- Use the artist name and song title in different orders
- Try using a direct YouTube URL instead of a search query

### 3. Wait and Retry

YouTube's anti-bot measures are sometimes temporary:
- Wait 10-15 minutes and try again
- The issue often resolves itself after a short period

## Advanced Solutions

### Enhanced Download Strategies

Shadowbox now automatically tries multiple download strategies when the first one fails:

1. **Enhanced Headers**: Uses realistic browser headers and user agents
2. **Alternative Extractor**: Uses Android client for extraction
3. **Web Client**: Uses web client with browser-like behavior
4. **Original Method**: Falls back to the original aria2c method

These strategies are applied automatically when you encounter download issues.

### Manual yt-dlp Commands

If you want to try downloading manually, here are some commands that often work:

```bash
# Strategy 1: Enhanced headers
yt-dlp --user-agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
       --add-header "Accept-Language:en-US,en;q=0.9" \
       -x --audio-format opus "ytsearch1:your search query"

# Strategy 2: Android client
yt-dlp --extractor-args "youtube:player_client=android" \
       --user-agent "Mozilla/5.0 (Linux; Android 11; SM-G973F) AppleWebKit/537.36" \
       -x --audio-format opus "ytsearch1:your search query"

# Strategy 3: Web client
yt-dlp --extractor-args "youtube:player_client=web" \
       -x --audio-format opus "ytsearch1:your search query"
```

## Prevention Tips

1. **Keep yt-dlp Updated**: YouTube frequently changes their systems, and yt-dlp updates regularly to counter these changes.

2. **Use Reasonable Download Rates**: Don't download too many songs in rapid succession.

3. **Vary Your Searches**: Use different search terms and don't repeat the same queries too frequently.

4. **Consider Using Direct URLs**: If you have specific YouTube URLs, use them instead of search queries.

## Alternative Sources

If YouTube continues to give you trouble, consider:

1. **Bandcamp**: Shadowbox supports Bandcamp downloads, which are generally more reliable
2. **Direct URLs**: Use specific YouTube video URLs instead of search queries
3. **Different Times**: Try downloading during off-peak hours

## Getting Help

If none of these solutions work:

1. Check the [yt-dlp GitHub issues](https://github.com/yt-dlp/yt-dlp/issues) for recent reports
2. Make sure you're using the latest version of both Shadowbox and yt-dlp
3. Try the manual yt-dlp commands above to isolate the issue
4. Consider that some videos may be region-locked or unavailable

## Technical Details

The "Sign in to confirm you're not a bot" error occurs when:
- YouTube detects automated download patterns
- The user agent or headers don't look like a real browser
- Too many requests are made in a short time period
- YouTube implements new anti-bot measures

Shadowbox's enhanced download system addresses these issues by:
- Using realistic browser user agents and headers
- Implementing multiple extraction strategies
- Adding delays between retry attempts
- Providing fallback methods when the primary approach fails

## Version Information

This enhanced download system was introduced in Shadowbox v1.0.1 to address the increasing frequency of YouTube anti-bot measures.