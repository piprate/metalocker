// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expiry

import (
	"errors"
	"time"
)

var (
	BadTime = time.Unix(666, 0)

	unitMap = map[string]uint64{
		"ms":  uint64(time.Millisecond),
		"s":   uint64(time.Second),
		"min": uint64(time.Minute),
		"h":   uint64(time.Hour),
	}
)

func FromNow(s string) time.Time {
	return FromDate(time.Now().UTC(), s)
}

func FromDate(base time.Time, s string) time.Time {
	offset, err := FromDateErr(base, s)
	if err != nil {
		return BadTime
	} else {
		return offset
	}
}

// FromDateErr parses a lease duration string.
// A lease duration string is a sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "10y", "1m10d", "1h30min", "300ms", "0" or "never".
// Valid time units are "y", "m", "d", "ms", "s", "min", "h".
// Fractions are not allowed for years, months and days.
// "never" or "0" means no expiry.
func FromDateErr(base time.Time, s string) (time.Time, error) { //nolint:gocyclo
	// ([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64

	// Special case: if all that is left is "0" or "never", return zero.
	if s == "0" || s == "never" {
		return time.Time{}, nil
	}
	if s == "" {
		return BadTime, errors.New("invalid offset " + quote(orig))
	}
	var years, months, days int
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return BadTime, errors.New("invalid offset " + quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return BadTime, errors.New("invalid offset " + quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return BadTime, errors.New("invalid offset " + quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return BadTime, errors.New("missing unit in offset " + quote(orig))
		}
		u := s[:i]
		s = s[i:]
		switch u {
		case "y":
			if f > 0 {
				return BadTime, errors.New("fraction not allowed in offset " + quote(orig))
			}
			years = int(v)
		case "m":
			if f > 0 {
				return BadTime, errors.New("fraction not allowed in offset " + quote(orig))
			}
			months = int(v)
		case "d":
			if f > 0 {
				return BadTime, errors.New("fraction not allowed in offset " + quote(orig))
			}
			days = int(v)
		default:
			unit, ok := unitMap[u]
			if !ok {
				return BadTime, errors.New("unknown unit " + quote(u) + " in offset " + quote(orig))
			}
			if v > 1<<63/unit {
				// overflow
				return BadTime, errors.New("invalid offset " + quote(orig))
			}
			v *= unit
			if f > 0 {
				// float64 is needed to be nanosecond accurate for fractions of hours.
				// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
				v += uint64(float64(f) * (float64(unit) / scale))
				if v > 1<<63 {
					// overflow
					return BadTime, errors.New("invalid offset " + quote(orig))
				}
			}
			d += v
			if d > 1<<63 {
				return BadTime, errors.New("invalid offset " + quote(orig))
			}
		}
	}

	if d > 1<<63-1 {
		return BadTime, errors.New("invalid offset " + quote(orig))
	}

	if years > 0 || months > 0 || days > 0 {
		base = base.AddDate(years, months, days)
	}

	if d > 0 {
		base = base.Add(time.Duration(d))
	}

	return base, nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x uint64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

// These are borrowed from unicode/utf8 and strconv and replicate behavior in
// that package, since we can't take a dependency on either.
const (
	lowerhex  = "0123456789abcdef"
	runeSelf  = 0x80
	runeError = '\uFFFD'
)

func quote(s string) string {
	buf := make([]byte, 1, len(s)+2) // slice will be at least len(s) + quotes
	buf[0] = '"'
	for i, c := range s {
		if c >= runeSelf || c < ' ' {
			// This means you are asking us to parse a time.Duration or
			// time.Location with unprintable or non-ASCII characters in it.
			// We don't expect to hit this case very often. We could try to
			// reproduce strconv.Quote's behavior with full fidelity but
			// given how rarely we expect to hit these edge cases, speed and
			// conciseness are better.
			var width int
			if c == runeError {
				width = 1
				if i+2 < len(s) && s[i:i+3] == string(runeError) {
					width = 3
				}
			} else {
				width = len(string(c))
			}
			for j := 0; j < width; j++ {
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[i+j]>>4], lowerhex[s[i+j]&0xF])
			}
		} else {
			if c == '"' || c == '\\' {
				buf = append(buf, '\\')
			}
			buf = append(buf, string(c)...)
		}
	}
	buf = append(buf, '"')
	return string(buf)
}
