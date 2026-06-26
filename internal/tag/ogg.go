package tag

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// This file implements the minimal Ogg container handling needed to rewrite the
// comment header of an Ogg stream (used for Opus). It parses pages, lets the
// caller replace the second packet (the comment header), and re-serialises the
// stream with corrected page sequence numbers and CRC checksums.

const oggCapture = "OggS"

// oggPage is a single parsed Ogg page.
type oggPage struct {
	headerType byte
	granule    uint64
	serial     uint32
	segTable   []byte
	body       []byte
}

// oggCRCTable is the lookup table for the Ogg CRC variant (poly 0x04c11db7,
// no input/output reflection, zero init and xorout).
var oggCRCTable = func() [256]uint32 {
	const poly = 0x04c11db7
	var t [256]uint32
	for i := 0; i < 256; i++ {
		crc := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
		t[i] = crc
	}
	return t
}()

func oggCRC(data []byte) uint32 {
	var crc uint32
	for _, b := range data {
		crc = (crc << 8) ^ oggCRCTable[byte(crc>>24)^b]
	}
	return crc
}

// parseOggPages parses every page in data in order.
func parseOggPages(data []byte) ([]oggPage, error) {
	var pages []oggPage
	pos := 0
	for pos < len(data) {
		if pos+27 > len(data) {
			return nil, errors.New("truncated ogg page header")
		}
		if string(data[pos:pos+4]) != oggCapture {
			return nil, fmt.Errorf("invalid ogg capture pattern at offset %d", pos)
		}
		headerType := data[pos+5]
		granule := binary.LittleEndian.Uint64(data[pos+6 : pos+14])
		serial := binary.LittleEndian.Uint32(data[pos+14 : pos+18])
		numSegments := int(data[pos+26])
		segStart := pos + 27
		if segStart+numSegments > len(data) {
			return nil, errors.New("truncated ogg segment table")
		}
		segTable := append([]byte(nil), data[segStart:segStart+numSegments]...)

		bodyLen := 0
		for _, s := range segTable {
			bodyLen += int(s)
		}
		bodyStart := segStart + numSegments
		if bodyStart+bodyLen > len(data) {
			return nil, errors.New("truncated ogg page body")
		}
		body := append([]byte(nil), data[bodyStart:bodyStart+bodyLen]...)

		pages = append(pages, oggPage{
			headerType: headerType,
			granule:    granule,
			serial:     serial,
			segTable:   segTable,
			body:       body,
		})
		pos = bodyStart + bodyLen
	}
	return pages, nil
}

// serialize renders the page with the given sequence number and a freshly
// computed CRC.
func (p oggPage) serialize(seq uint32) []byte {
	buf := make([]byte, 27+len(p.segTable)+len(p.body))
	copy(buf[0:4], oggCapture)
	buf[4] = 0 // version
	buf[5] = p.headerType
	binary.LittleEndian.PutUint64(buf[6:14], p.granule)
	binary.LittleEndian.PutUint32(buf[14:18], p.serial)
	binary.LittleEndian.PutUint32(buf[18:22], seq)
	// CRC field (22:26) left zero for computation.
	buf[26] = byte(len(p.segTable))
	copy(buf[27:], p.segTable)
	copy(buf[27+len(p.segTable):], p.body)

	crc := oggCRC(buf)
	binary.LittleEndian.PutUint32(buf[22:26], crc)
	return buf
}

// packetToPages splits a single packet into one or more pages for the given
// serial, marking continuation pages appropriately. All pages use the supplied
// granule and base header type (continuation flag is OR-ed in as needed).
func packetToPages(packet []byte, serial uint32, granule uint64, baseHeaderType byte) []oggPage {
	// Build the full lacing table for the packet.
	var lacing []byte
	remaining := len(packet)
	for {
		if remaining >= 255 {
			lacing = append(lacing, 255)
			remaining -= 255
		} else {
			lacing = append(lacing, byte(remaining))
			break
		}
	}
	// A packet whose length is an exact multiple of 255 needs a terminating 0.
	if len(packet) > 0 && len(packet)%255 == 0 {
		lacing = append(lacing, 0)
	}

	var pages []oggPage
	bodyPos := 0
	first := true
	for len(lacing) > 0 {
		n := len(lacing)
		if n > 255 {
			n = 255
		}
		chunk := lacing[:n]
		lacing = lacing[n:]

		chunkLen := 0
		for _, s := range chunk {
			chunkLen += int(s)
		}

		ht := baseHeaderType
		if !first {
			ht |= 0x01 // continued packet
		}
		pages = append(pages, oggPage{
			headerType: ht,
			granule:    granule,
			serial:     serial,
			segTable:   append([]byte(nil), chunk...),
			body:       append([]byte(nil), packet[bodyPos:bodyPos+chunkLen]...),
		})
		bodyPos += chunkLen
		first = false
	}
	return pages
}

// packetSpan reports how many leading pages (starting at idx) make up a single
// packet, returning the packet bytes and the number of pages consumed. It
// returns ok=false if the packet shares a page with a following packet (which
// this minimal rewriter does not handle).
func packetSpan(pages []oggPage, idx int) (packet []byte, consumed int, ok bool) {
	for i := idx; i < len(pages); i++ {
		p := pages[i]
		packet = append(packet, p.body...)
		// If the final segment of this page is < 255, the packet ends here.
		if len(p.segTable) == 0 {
			return nil, 0, false
		}
		last := p.segTable[len(p.segTable)-1]
		// Ensure every segment in this page belongs to our packet: only the
		// final segment may be < 255. If an earlier segment is < 255, another
		// packet starts within this page and we bail out.
		for j := 0; j < len(p.segTable)-1; j++ {
			if p.segTable[j] < 255 {
				return nil, 0, false
			}
		}
		if last < 255 {
			return packet, i - idx + 1, true
		}
	}
	return nil, 0, false
}
