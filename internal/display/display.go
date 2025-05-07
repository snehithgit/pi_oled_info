package display

import (
	"fmt"
	"strings"

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
	if d.dev != nil {
		d.dev.Clear()
		for i := range d.lines {
			d.lines[i] = ""
		}
	}
}

// Update refreshes the display with current line contents
func (d *Display) Update() error {
	if d.dev == nil {
		return fmt.Errorf("display not initialized")
	}

	// Clear the display first
	if err := d.dev.Clear(); err != nil {
		return fmt.Errorf("failed to clear display: %v", err)
	}

	// Draw each line
	for i, line := range d.lines {
		if line != "" {
			y := i * charHeight // Line position
			if err := d.dev.DrawString(0, y, strings.TrimSpace(line), nil); err != nil {
				return fmt.Errorf("failed to draw line %d: %v", i, err)
			}
		}
	}

	return nil
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
