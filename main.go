package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	bcdPlusCodes = "0123456789 -."
	bcdPlusSP    = 10 // BCD Plus space code
)

type BCDPlusBytes []byte

func (b BCDPlusBytes) Decode(trim bool) ([]byte, error) {
	lSrc, nCodes := len(b), len(bcdPlusCodes)
	dest := make([]byte, lSrc*2)
	for i, sb := range b {
		for j := 0; j < 2; j++ {
			v := (int(sb) >> ((1 - j) * 4)) & 0x0f
			if v >= nCodes {
				loc := "upper"
				if j > 0 {
					loc = "lower"
				}
				return nil, fmt.Errorf("invalide BCD Plus code (%d) in %s nibble of byte #%d of %q", v, loc, i, b)
			}
			dest[i*2+j] = bcdPlusCodes[v]
		}
	}
	if trim {
		return bytes.TrimRight(dest, " "), nil
	}
	return dest, nil
}

func (b *BCDPlusBytes) Encode(src []byte) (bool, error) {
	lSrc := len(src)
	lDest := (lSrc + 1) / 2
	*b = make([]byte, lDest)
	for i := 0; i < lDest; i++ {
		code := 0
		for j := 0; j < 2; j++ {
			sIdx := 2*i + j
			var idx int
			if sIdx < lSrc {
				sb := src[sIdx]
				idx = strings.IndexByte(bcdPlusCodes, sb)
				if idx < 0 {
					return false, fmt.Errorf("invalid char %q for BCD Plus encoding at index %d of %q", sb, sIdx, src)
				}
			} else { // no more src, padd with SP
				idx = bcdPlusSP
			}
			code = code*16 + idx
		}
		(*b)[i] = byte(code)
	}
	return lSrc%2 != 0, nil
}

const (
	firstPacked6BitAscii = 0x20 // ASCII space code
	lastPacked6BitAscii  = 0x5f // ASCII '_' code
)

type Packed6BitAsciiBytes []byte

func (p Packed6BitAsciiBytes) Decode(trim bool) ([]byte, error) {
	lSrc := len(p)
	lDest := (lSrc/3)*4 + (lSrc % 3)
	dest := make([]byte, lDest)
	var remain byte
	j := 0
	for i, sb := range p {
		var v byte
		switch i % 3 {
		case 0:
			v = sb & 0x3f
			remain = (sb >> 6) & 0x03
		case 1:
			v = (sb&0x0f)<<2 | remain
			remain = (sb >> 4) & 0x0f
		case 2:
			v = (sb&0x03)<<4 | remain
			dest[j] = v + firstPacked6BitAscii
			j++
			v = (sb >> 2) & 0x3f
			remain = 0
		}
		dest[j] = v + firstPacked6BitAscii
		j++
	}
	if trim {
		return bytes.TrimRight(dest, " "), nil
	}
	return dest, nil
}

func (p *Packed6BitAsciiBytes) Encode(src []byte) error {
	lSrc := len(src)
	lDest := (lSrc/4)*3 + (lSrc % 4)
	*p = make([]byte, lDest)
	j := 0
	var acc byte
	for i, sb := range src {
		if sb < firstPacked6BitAscii || sb > lastPacked6BitAscii {
			return fmt.Errorf("invalid char %q for Packed 6-bit ASCII encoding at index %d of %q", sb, i, src)
		}
		sb -= firstPacked6BitAscii
		switch i % 4 {
		case 0:
			acc = sb & 0x3f // 6 bits
		case 1:
			(*p)[j] = (sb&0x03)<<6 | acc
			acc = (sb >> 2) & 0x0f // 4 bits
			j++
		case 2:
			(*p)[j] = (sb&0x0f)<<4 | acc
			acc = (sb >> 4) & 0x03 // 2 bits
			j++
		case 3:
			(*p)[j] = (sb&0x3f)<<2 | acc
			acc = 0
			j++
		}
	}
	if acc != 0 {
		(*p)[j] = acc
	}
	return nil
}

func main() {
	s := "123-456-7.890"
	var x BCDPlusBytes
	fmt.Println(x.Encode([]byte(s)))
	fmt.Println(hex.Dump(x))
	sb, err := x.Decode(false)
	fmt.Printf("%q err=%v\n", sb, err)

	t := "IPMITOOL 12"
	var p Packed6BitAsciiBytes
	fmt.Println(p.Encode([]byte(t)))
	fmt.Println(hex.Dump(p))
	sb, err = p.Decode(true)
	fmt.Printf("%q err=%v\n", sb, err)
}
