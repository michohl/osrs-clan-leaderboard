package skills

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/michohl/osrs-clan-leaderboard/jet_schemas/model"
	"github.com/michohl/osrs-clan-leaderboard/types"
	"golang.org/x/image/font"
)

const (
	SKILLS_TEMPLATE_PATH = "skills/resources/skills_tab_template.png"
	FONT_PATH            = "skills/resources/runescape.ttf"
)

var AccountTypeShorthands = map[string]string{
	"hardcore_ironman":       "HCIM",
	"ultimate_ironman":       "UIM",
	"ironman":                "IM",
	"main":                   "Main",
	"unranked_group_ironman": "Unranked GIM",
	"group_ironman":          "GIM",
	"hardcore_group_ironman": "HCGIM",
}

// RenderSkillsPage takes the template image, font bytes, and user hiscores,
// then returns a PNG encoded byte slice ready to be sent in a Discord message.
func RenderSkillsPage(user model.Users, hiscores *types.Hiscores) ([]byte, error) {
	// Load template image
	imgFile, err := os.Open(SKILLS_TEMPLATE_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to open template image: %w", err)
	}
	defer imgFile.Close()

	srcImg, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode template image: %w", err)
	}

	// Convert to *image.RGBA so we can draw on it
	bounds := srcImg.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, srcImg, bounds.Min, draw.Src)

	// Load TrueType font (runescape.ttf)
	fontBytes, err := os.ReadFile(FONT_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %w", err)
	}

	fnt, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse truetype font: %w", err)
	}

	cYellow := color.RGBA{255, 255, 0, 255}
	cBlack := color.RGBA{0, 0, 0, 255}

	// Grid geometry: OSRS skill tab columns and row baselines
	// This is the bottom left most pixel of the rendered string
	colXs := []int{42, 105, 168}
	rowYs := []int{42, 72, 102, 132, 162, 192, 222, 252}

	// OSRS lists skills vertically per column (8 rows per column, 3 columns total)
	columnSkills := [3][]string{
		{"Attack", "Strength", "Defence", "Ranged", "Prayer", "Magic", "Runecraft", "Construction"},
		{"Hitpoints", "Agility", "Herblore", "Thieving", "Crafting", "Fletching", "Slayer", "Hunter"},
		{"Mining", "Smithing", "Fishing", "Cooking", "Firemaking", "Woodcutting", "Farming", "Sailing"},
	}

	// Helper to draw text with a 1px black drop-shadow at font size 10
	drawTextWithShadow := func(dst *image.RGBA, text string, x, y int, size float64, col color.Color) {
		drawText(dst, fnt, text, x+1, y+1, size, cBlack)
		drawText(dst, fnt, text, x, y, size, col)
	}

	// Render each skill's current and base levels on either side of the slash
	fontSize := 12.0

	for cIdx, col := range columnSkills {
		for rIdx, skillName := range col {
			skill := hiscores.GetSkill(skillName)
			if skill == nil {
				continue
			}

			currentStr := fmt.Sprintf("%d", skill.Level)
			baseStr := fmt.Sprintf("%d", skill.Level)

			x := colXs[cIdx]
			y := rowYs[rIdx]

			// If the level is a single digit then we need to
			// bump it over slightly to the right to make it align well
			lMod := 0
			rMod := 0
			if skill.Level < 10 {
				lMod = 4
				rMod = 2
			}

			// Left side of the slash
			drawTextWithShadow(canvas, currentStr, x+lMod, y, fontSize, cYellow)

			// Right side of the slash
			drawTextWithShadow(canvas, baseStr, x+rMod+12, y+13, fontSize, cYellow)
		}
	}

	// Find center of canvas
	imgWidth := bounds.Dx()
	xCenter := imgWidth / 2

	titleCard := fmt.Sprintf("%s (%s) - %s", user.OsrsUsername, AccountTypeShorthands[user.OsrsAccountType], time.Now().Format("01/02/2006"))
	// Centered in the top banner area
	drawTextWithShadow(canvas, titleCard, xCenter-(measureTextWidth(fnt, titleCard, fontSize)/2), 20, fontSize, cYellow)

	totalLevel := hiscores.GetSkill("Overall").Level
	totalStr := fmt.Sprintf("%d", totalLevel)
	// Centered in the bottom banner area
	drawTextWithShadow(canvas, totalStr, 122, 282, fontSize, cYellow)

	// Encode canvas to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, fmt.Errorf("failed to encode result image: %w", err)
	}

	if buf.Len() == 0 {
		return nil, fmt.Errorf("Rendered an image of zero bytes")
	}

	return buf.Bytes(), nil
}

// Helper function to render string via freetype onto an RGBA image
func drawText(canvas *image.RGBA, fnt *truetype.Font, text string, x, y int, size float64, col color.Color) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fnt)
	c.SetFontSize(size)
	c.SetClip(canvas.Bounds())
	c.SetDst(canvas)
	c.SetSrc(image.NewUniform(col))

	pt := freetype.Pt(x, y)
	_, _ = c.DrawString(text, pt)
}

func measureTextWidth(fnt *truetype.Font, text string, size float64) int {
	// Create a font face from the truetype font
	face := truetype.NewFace(fnt, &truetype.Options{
		Size: size,
		DPI:  72,
	})

	// MeasureString returns fixed.26_6 fixed-point pixels
	widthInt26_6 := font.MeasureString(face, text)

	// Convert fixed-point to integer pixels (rounding up or down)
	return int(widthInt26_6.Round())
}
