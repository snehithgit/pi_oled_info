// Updated display.go to implement anti-flicker display updates

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
	dev           *ssd1306.Dev
	lines         [maxLines]string
	prevLines     [maxLines]string      // Track previous state for each line
	charPositions [maxLines][maxChars]bool // Track which character positions changed
	i2cBus        i2c.BusCloser
	img           *image.RGBA // Keep a buffer of the current display state
	initialized   bool        // Track if the display has been fully drawn once
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

	// Create empty image buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	return &Display{
		dev:         dev,
		i2cBus:      bus,
		img:         img,
		initialized: false,
	}, nil
}

// WriteLine sets the content for a specific line
func (d *Display) WriteLine(line int, content string) {
	if line < 0 || line >= maxLines {
		return
	}

	// Truncate content if needed
	if len(content) > maxChars {
		content = content[:maxChars]
	}
	
	// If line content has changed, update and mark characters that need redrawing
	if content != d.lines[line] {
		// Compare each character position to determine which ones changed
		for i := 0; i < maxChars; i++ {
			var prevChar, newChar byte
			
			if i < len(d.lines[line]) {
				prevChar = d.lines[line][i]
			} else {
				prevChar = ' '
			}
			
			if i < len(content) {
				newChar = content[i]
			} else {
				newChar = ' '
			}
			
			// Mark character position for update if character has changed
			d.charPositions[line][i] = prevChar != newChar
		}
		
		// Update the line content
		d.lines[line] = content
	}
}

// Clear clears all lines on the display
func (d *Display) Clear() {
	for i := range d.lines {
		if d.lines[i] != "" {
			d.lines[i] = ""
			// Mark all characters as needing update
			for j := range d.charPositions[i] {
				d.charPositions[i][j] = true
			}
		}
	}
	
	// Only perform a full clear if we need to
	if d.initialized {
		d.redrawChangedChars()
	} else {
		// If not initialized, do a full clear
		draw.Draw(d.img, d.img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
		d.dev.Draw(d.img.Bounds(), d.img, image.Point{})
	}
}

// Update refreshes the display with current line contents
func (d *Display) Update() error {
	if d.dev == nil {
		return fmt.Errorf("display not initialized")
	}
	
	if !d.initialized {
		// First time: full draw of everything
		if err := d.fullDraw(); err != nil {
			return err
		}
		d.initialized = true
		// Save current state as previous
		for i := range d.lines {
			d.prevLines[i] = d.lines[i]
		}
		return nil
	}
	
	// Check if any character has changed
	needsUpdate := false
	for _, line := range d.charPositions {
		for _, changed := range line {
			if changed {
				needsUpdate = true
				break
			}
		}
		if needsUpdate {
			break
		}
	}
	
	if !needsUpdate {
		return nil // Nothing changed, no need to update
	}
	
	// Update only the changed characters
	return d.redrawChangedChars()
}

// fullDraw performs a complete redraw of the display
func (d *Display) fullDraw() error {
	// Create a black image buffer
	draw.Draw(d.img, d.img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

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
			Dst:  d.img,
			Src:  image.White,
			Face: face,
			Dot:  dot,
		}
		drawer.DrawString(strings.TrimSpace(line))
	}

	// Push image to OLED display
	return d.dev.Draw(d.img.Bounds(), d.img, image.Point{})
}

// redrawChangedChars updates only the characters that have changed
func (d *Display) redrawChangedChars() error {
	// Use a basic font 
	face := basicfont.Face7x13
	
	// Check each position that needs updating
	for lineIdx, line := range d.charPositions {
		for charIdx, needsUpdate := range line {
			if !needsUpdate {
				continue
			}
			
			// Calculate the position for this character
			x := charIdx * charWidth
			y := lineIdx * charHeight
			
			// Create a small rect for just this character
			charRect := image.Rect(x, y, x+charWidth, y+charHeight)
			
			// Clear this character position
			draw.Draw(d.img, charRect, &image.Uniform{color.Black}, image.Point{}, draw.Src)
			
			// If there's a character to draw at this position, draw it
			if lineIdx < len(d.lines) && charIdx < len(d.lines[lineIdx]) {
				char := string(d.lines[lineIdx][charIdx])
				dot := fixed.Point26_6{
					X: fixed.I(x),
					Y: fixed.I(y + charHeight - 3), // Adjust vertical spacing
				}
				drawer := &font.Drawer{
					Dst:  d.img,
					Src:  image.White,
					Face: face,
					Dot:  dot,
				}
				drawer.DrawString(char)
			}
			
			// Update only this part of the display
			if err := d.dev.Draw(charRect, d.img, image.Point{X: x, Y: y}); err != nil {
				return err
			}
			
			// Mark as updated
			d.charPositions[lineIdx][charIdx] = false
		}
	}
	
	// Save current state as previous
	for i := range d.lines {
		d.prevLines[i] = d.lines[i]
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
