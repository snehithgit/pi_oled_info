package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

const (
	width      = 128
	height     = 64
	charWidth  = 8
	charHeight = 16
	maxChars   = 16 // Maximum characters per line (width / charWidth)
	maxLines   = 4  // Maximum lines (height / charHeight)
)

// Display handles the OLED screen operations
type Display struct {
	dev    *ssd1306.Dev
	lines  [maxLines]string
	i2cBus i2c.BusCloser
}

// New creates and initializes a new Display
func New() (*Display, error) {
	// Initialize periph.io
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph: %v", err)
	}

	// Open I2C bus
	bus, err := i2creg.Open("")
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus: %v", err)
	}

	// Initialize the SSD1306 display (default address 0x3C)
	dev, err := ssd1306.NewI2C(bus, &ssd1306.Opts{
		W:       width,
		H:       height,
		Rotated: false,
	})
	if err != nil {
		bus.Close()
		return nil, fmt.Errorf("failed to initialize SSD1306: %v", err)
	}

	return &Display{
		dev:    dev,
		i2cBus: bus,
	}, nil
}

// WriteLine sets the content for a specific line
func (d *Display) WriteLine(line int, content string) {
	if line < 0 || line >= maxLines {
		return
	}

	// Truncate content if needed and pad to have consistent display
	if len(content) > maxChars {
		content = content[:maxChars]
	}
	d.lines[line] = content
}

// Clear clears all lines on the display
func (d *Display) Clear() {
	for i := range d.lines {
		d.lines[i] = ""
	}
	d.Update() // Redraw blank screen
}

// Update refreshes the display with current line contents
func (d *Display) Update() error {
	if d.dev == nil {
		return fmt.Errorf("display not initialized")
	}

	// Create a black image buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// Use a basic font to draw each line
	face := basicfont.Face7x13
	for i, line := range d.lines {
		if line == "" {
			continue
		}
		dot := fixed.Point26_6{
			X: fixed.I(0),
			Y: fixed.I((i+1)*charHeight - 3), // Adjust vertical spacing
		}
		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.White,
			Face: face,
			Dot:  dot,
		}
		drawer.DrawString(strings.TrimSpace(line))
	}

	// Push image to OLED display
	return d.dev.Draw(img.Bounds(), img, image.Point{})
}

// Close shuts down the display properly
func (d *Display) Close() {
	if d.dev != nil {
		d.dev.Halt()
	}
	if d.i2cBus != nil {
		d.i2cBus.Close()
	}
}
