package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Color map[string]string

func (c Color) String() string { return c["hex"] }

type rgba struct {
	r, g, b uint8
	a       float64
}

var funcColorRe = regexp.MustCompile(`^(rgba?|hsla?)\(([^()]*)\)$`)

func parseColor(s string) (rgba, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return rgba{}, false
	}
	if strings.HasPrefix(s, "#") {
		return parseHex(s[1:])
	}
	if m := funcColorRe.FindStringSubmatch(strings.ToLower(s)); m != nil {
		return parseFunc(m[1], m[2])
	}
	return rgba{}, false
}

func parseHex(h string) (rgba, bool) {
	for _, r := range h {
		if !strings.ContainsRune("0123456789abcdefABCDEF", r) {
			return rgba{}, false
		}
	}

	if len(h) == 3 || len(h) == 4 {
		var expanded strings.Builder
		for _, r := range h {
			expanded.WriteRune(r)
			expanded.WriteRune(r)
		}
		h = expanded.String()
	}
	if len(h) != 6 && len(h) != 8 {
		return rgba{}, false
	}

	n, err := strconv.ParseUint(h, 16, 64)
	if err != nil {
		return rgba{}, false
	}
	if len(h) == 6 {
		return rgba{r: uint8(n >> 16), g: uint8(n >> 8), b: uint8(n), a: 1}, true
	}
	return rgba{
		r: uint8(n >> 24),
		g: uint8(n >> 16),
		b: uint8(n >> 8),
		a: float64(uint8(n)) / 255,
	}, true
}

func parseFunc(name, args string) (rgba, bool) {
	fields := strings.FieldsFunc(args, func(r rune) bool { return r == ',' || r == '/' })
	if len(fields) != 3 && len(fields) != 4 {
		return rgba{}, false
	}

	nums := make([]float64, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(f), "%"), 64)
		if err != nil {
			return rgba{}, false
		}
		nums[i] = v
	}

	a := 1.0
	if len(nums) == 4 {
		a = clamp(nums[3], 0, 1)
	}

	switch name {
	case "rgb", "rgba":
		return rgba{
			r: uint8(math.Round(clamp(nums[0], 0, 255))),
			g: uint8(math.Round(clamp(nums[1], 0, 255))),
			b: uint8(math.Round(clamp(nums[2], 0, 255))),
			a: a,
		}, true
	case "hsl", "hsla":
		r, g, b := hslToRGB(nums[0], clamp(nums[1], 0, 100)/100, clamp(nums[2], 0, 100)/100)
		return rgba{r: r, g: g, b: b, a: a}, true
	}
	return rgba{}, false
}

func (c rgba) hsl() (h, s, l float64) {
	r, g, b := float64(c.r)/255, float64(c.g)/255, float64(c.b)/255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2

	if max == min {
		return 0, 0, l * 100
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}

	return h * 60, s * 100, l * 100
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360

	if s == 0 {
		v := uint8(math.Round(l * 255))
		return v, v, v
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	return hueToByte(p, q, h+1.0/3), hueToByte(p, q, h), hueToByte(p, q, h-1.0/3)
}

func hueToByte(p, q, t float64) uint8 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}

	switch {
	case t < 1.0/6:
		p += (q - p) * 6 * t
	case t < 1.0/2:
		p = q
	case t < 2.0/3:
		p += (q - p) * (2.0/3 - t) * 6
	}
	return uint8(math.Round(p * 255))
}

func newColor(c rgba) Color {
	hex := fmt.Sprintf("%02x%02x%02x", c.r, c.g, c.b)
	alphaHex := fmt.Sprintf("%02x", uint8(math.Round(c.a*255)))
	h, s, l := c.hsl()
	hs, ss, ls := fmtFloat(h, 1), fmtFloat(s, 1), fmtFloat(l, 1)
	as := fmtFloat(c.a, 2)

	return Color{
		"hex":                "#" + hex,
		"hex_stripped":       hex,
		"hex_alpha":          "#" + hex + alphaHex,
		"hex_alpha_stripped": hex + alphaHex,
		"alpha_hex":          "#" + alphaHex + hex,
		"alpha_hex_stripped": alphaHex + hex,
		"rgb":                fmt.Sprintf("rgb(%d, %d, %d)", c.r, c.g, c.b),
		"rgba":               fmt.Sprintf("rgba(%d, %d, %d, %s)", c.r, c.g, c.b, as),
		"hsl":                fmt.Sprintf("hsl(%s, %s%%, %s%%)", hs, ss, ls),
		"hsla":               fmt.Sprintf("hsla(%s, %s%%, %s%%, %s)", hs, ss, ls, as),
		"red":                strconv.Itoa(int(c.r)),
		"green":              strconv.Itoa(int(c.g)),
		"blue":               strconv.Itoa(int(c.b)),
		"alpha":              as,
		"hue":                hs,
		"saturation":         ss + "%",
		"lightness":          ls + "%",
	}
}

func fmtFloat(v float64, prec int) string {
	s := strconv.FormatFloat(v, 'f', prec, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(math.Max(v, lo), hi)
}

func resolveColors(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = resolveColors(val)
		}
	case []any:
		for i, val := range t {
			t[i] = resolveColors(val)
		}
	case string:
		if c, ok := parseColor(t); ok {
			return newColor(c)
		}
	}
	return v
}
