package main

import (
	"image/color"
	"strconv"
	"strings"
)

// parseConfigColor разбирает цвет из конфига: либо номер ANSI (0-255), либо
// hex "#RRGGBB". Возвращает запасной цвет, если значение пустое или невалидное.
func parseConfigColor(value string) color.NRGBA {
	fallback := color.NRGBA{R: 0x5f, G: 0xaf, B: 0xff, A: 0xff}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasPrefix(value, "#") {
		hex := strings.TrimPrefix(value, "#")
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) != 6 {
			return fallback
		}
		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return fallback
		}
		return color.NRGBA{
			R: uint8(v >> 16),
			G: uint8(v >> 8),
			B: uint8(v),
			A: 0xff,
		}
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || n > 255 {
		return fallback
	}
	return ansiToNRGBA(n)
}

// ansiToNRGBA конвертирует номер ANSI (0-255) в RGB по стандартной палитре
// xterm: 16 базовых цветов + куб 6x6x6 + градации серого.
func ansiToNRGBA(n int) color.NRGBA {
	base := [16][3]uint8{
		{0x00, 0x00, 0x00}, {0x80, 0x00, 0x00}, {0x00, 0x80, 0x00}, {0x80, 0x80, 0x00},
		{0x00, 0x00, 0x80}, {0x80, 0x00, 0x80}, {0x00, 0x80, 0x80}, {0xc0, 0xc0, 0xc0},
		{0x80, 0x80, 0x80}, {0xff, 0x00, 0x00}, {0x00, 0xff, 0x00}, {0xff, 0xff, 0x00},
		{0x00, 0x00, 0xff}, {0xff, 0x00, 0xff}, {0x00, 0xff, 0xff}, {0xff, 0xff, 0xff},
	}
	if n < 16 {
		return color.NRGBA{base[n][0], base[n][1], base[n][2], 0xff}
	}
	if n < 232 {
		n -= 16
		levels := [6]uint8{0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff}
		return color.NRGBA{
			levels[(n/36)%6],
			levels[(n/6)%6],
			levels[n%6],
			0xff,
		}
	}
	g := uint8(8 + (n-232)*10)
	return color.NRGBA{g, g, g, 0xff}
}