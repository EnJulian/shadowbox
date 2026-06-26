// Command shadowbox is a music acquisition CLI that downloads audio from
// YouTube and Bandcamp, converts it, and injects metadata, cover art, and lyrics.
package main

import (
	"os"

	"github.com/EnJulian/shadowbox/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
