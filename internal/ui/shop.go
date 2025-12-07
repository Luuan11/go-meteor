package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
)

// Shop UI constants
const (
	shopDescriptionOffsetY = 38
	shopPreviewOffsetX     = 250
	shopPreviewOffsetY     = 45
	shopMaxTextOffsetX     = 230
)

// Shop UI colors
var (
	colorUpgradePreview = color.RGBA{100, 255, 150, 255}
	colorMaxLevel       = color.RGBA{255, 215, 0, 255}
	colorDescription    = color.RGBA{255, 200, 100, 255}
	colorSelectedItemBg = color.RGBA{60, 60, 100, 230}
	colorUpgradeSuccess = color.RGBA{0, 200, 0, 230}
	colorHintBackground = color.RGBA{20, 20, 40, 230}
)

type ShopAction int

const (
	ShopActionNone ShopAction = iota
	ShopActionClose
	ShopActionUpgrade
)

type ShopItem struct {
	PowerType   string
	Name        string
	Icon        *ebiten.Image
	Level       int
	MaxLevel    int
	NextCost    int
	BonusPerLvl int
	IsSpecial   bool
	Description string
}

type Shop struct {
	Items           []ShopItem
	selectedIndex   int
	scrollOffset    int
	maxVisibleItems int
	coins           int
	action          ShopAction
	upgradeType     string
	showUpgradeMsg  bool
	msgTimer        int
}

func NewShop() *Shop {
	return &Shop{
		Items:           make([]ShopItem, 0),
		selectedIndex:   0,
		scrollOffset:    0,
		maxVisibleItems: 6,
		action:          ShopActionNone,
	}
}

func (s *Shop) SetProgress(progress *systems.PlayerProgress) {
	if progress == nil {
		return
	}

	s.coins = progress.Coins
	s.Items = []ShopItem{
		{PowerType: "superpower", Name: "Super Shot", Icon: assets.SuperPowerSprite, Level: progress.GetUpgradeLevel("superpower"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "shield", Name: "Shield", Icon: assets.ShieldPowerUpSprite, Level: progress.GetUpgradeLevel("shield"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "slowmotion", Name: "Slow Motion", Icon: assets.ClockPowerUpSprite, Level: progress.GetUpgradeLevel("slowmotion"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "laser", Name: "Laser Beam", Icon: assets.LaserPowerUpSprite, Level: progress.GetUpgradeLevel("laser"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "nuke", Name: "Nuke", Icon: assets.NukePowerUpSprite, Level: progress.GetUpgradeLevel("nuke"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "multiplier", Name: "Multiplier", Icon: assets.MultiplierPowerUpSprite, Level: progress.GetUpgradeLevel("multiplier"), MaxLevel: 5, BonusPerLvl: 2, IsSpecial: false},
		{PowerType: "coinmagnet", Name: "Coin Magnet", Icon: assets.CoinSprite, Level: progress.GetUpgradeLevel("coinmagnet"), MaxLevel: 1, BonusPerLvl: 0, IsSpecial: true},
		{PowerType: "doublecoins", Name: "Double Coins", Icon: assets.CoinSprite, Level: progress.GetUpgradeLevel("doublecoins"), MaxLevel: 1, BonusPerLvl: 0, IsSpecial: true},
		{PowerType: "startboost", Name: "Start with Boost", Icon: assets.SuperPowerSprite, Level: progress.GetUpgradeLevel("startboost"), MaxLevel: 1, BonusPerLvl: 0, IsSpecial: true},
	}

	for i := range s.Items {
		if s.Items[i].IsSpecial {
			switch s.Items[i].PowerType {
			case "coinmagnet":
				s.Items[i].NextCost = 500
			case "doublecoins":
				s.Items[i].NextCost = 250
			case "startboost":
				s.Items[i].NextCost = 250
			}
			if s.Items[i].Level >= s.Items[i].MaxLevel {
				s.Items[i].NextCost = 0
			}
		} else {
			s.Items[i].NextCost = s.calculateCost(s.Items[i].Level, s.Items[i].IsSpecial)
		}
	}
}

func (s *Shop) calculateCost(level int, isSpecial bool) int {
	if isSpecial {
		specialCosts := []int{100, 250, 500}
		if level >= len(specialCosts) {
			return 0
		}
		return specialCosts[level]
	}

	costs := []int{25, 50, 100, 200, 400}
	if level >= len(costs) {
		return 0
	}
	return costs[level]
}

func (s *Shop) Update() ShopAction {
	s.action = ShopActionNone

	if s.msgTimer > 0 {
		s.msgTimer--
		if s.msgTimer == 0 {
			s.showUpgradeMsg = false
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.action = ShopActionClose
		return s.action
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selectedIndex--
		if s.selectedIndex < 0 {
			s.selectedIndex = len(s.Items) - 1
		}
		s.updateScroll()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selectedIndex++
		if s.selectedIndex >= len(s.Items) {
			s.selectedIndex = 0
		}
		s.updateScroll()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if s.selectedIndex >= 0 && s.selectedIndex < len(s.Items) {
			item := s.Items[s.selectedIndex]
			if item.Level < item.MaxLevel && s.coins >= item.NextCost {
				s.action = ShopActionUpgrade
				s.upgradeType = item.PowerType
				s.showUpgradeMsg = true
				s.msgTimer = 60
				return s.action
			}
		}
	}

	return s.action
}

func (s *Shop) updateScroll() {
	if s.selectedIndex < s.scrollOffset {
		s.scrollOffset = s.selectedIndex
	} else if s.selectedIndex >= s.scrollOffset+s.maxVisibleItems {
		s.scrollOffset = s.selectedIndex - s.maxVisibleItems + 1
	}
}

func (s *Shop) Draw(screen *ebiten.Image) {
	overlay := ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 220})
	screen.DrawImage(overlay, nil)

	titleText := "SHOP"
	titleX := (config.ScreenWidth - measureText(titleText, assets.FontUi)) / 2
	drawText(screen, titleText, assets.FontUi, titleX, 40, color.RGBA{255, 215, 0, 255})

	coinOp := &ebiten.DrawImageOptions{}
	coinOp.GeoM.Scale(0.8, 0.8)
	coinOp.GeoM.Translate(float64(config.ScreenWidth/2-25), 55)
	screen.DrawImage(assets.CoinSprite, coinOp)
	coinsText := fmt.Sprintf("%d", s.coins)
	coinsX := (config.ScreenWidth - measureText(coinsText, assets.FontSmall)) / 2
	drawText(screen, coinsText, assets.FontSmall, coinsX+25, 75, color.White)

	startY := 95
	itemHeight := 62

	endIndex := s.scrollOffset + s.maxVisibleItems
	if endIndex > len(s.Items) {
		endIndex = len(s.Items)
	}

	for i := s.scrollOffset; i < endIndex; i++ {
		item := s.Items[i]
		displayIndex := i - s.scrollOffset
		y := startY + displayIndex*itemHeight

		bgColor := color.RGBA{30, 30, 50, 200}
		borderColor := color.RGBA{60, 60, 90, 255}
		if i == s.selectedIndex {
			bgColor = colorSelectedItemBg
			borderColor = color.RGBA{120, 120, 180, 255}
		}

		border := ebiten.NewImage(config.ScreenWidth-30, itemHeight-4)
		border.Fill(borderColor)
		borderOp := &ebiten.DrawImageOptions{}
		borderOp.GeoM.Translate(15, float64(y))
		screen.DrawImage(border, borderOp)

		bg := ebiten.NewImage(config.ScreenWidth-36, itemHeight-10)
		bg.Fill(bgColor)
		bgOp := &ebiten.DrawImageOptions{}
		bgOp.GeoM.Translate(18, float64(y+3))
		screen.DrawImage(bg, bgOp)

		iconOp := &ebiten.DrawImageOptions{}
		iconOp.GeoM.Scale(0.65, 0.65)
		iconOp.GeoM.Translate(25, float64(y+12))
		screen.DrawImage(item.Icon, iconOp)

		nameText := item.Name
		drawText(screen, nameText, assets.FontSmall, 75, y+20, color.White)

		if item.IsSpecial {
			if item.Description != "" {
				drawText(screen, item.Description, assets.FontSmall, 75, y+shopDescriptionOffsetY, colorDescription)
			}
		} else {
			levelText := fmt.Sprintf("Level %d/%d", item.Level, item.MaxLevel)
			drawText(screen, levelText, assets.FontSmall, 75, y+45, color.RGBA{180, 180, 220, 255})

			if item.Level < item.MaxLevel {
				currentDuration := 10 + (item.Level * item.BonusPerLvl)
				nextDuration := 10 + ((item.Level + 1) * item.BonusPerLvl)
				previewText := fmt.Sprintf("%ds - %ds", currentDuration, nextDuration)
				drawText(screen, previewText, assets.FontSmall, shopPreviewOffsetX, y+shopPreviewOffsetY, colorUpgradePreview)
			} else {
				maxText := "MAX"
				drawText(screen, maxText, assets.FontSmall, shopMaxTextOffsetX, y+shopPreviewOffsetY, colorMaxLevel)
			}
		}

		if item.Level >= item.MaxLevel {
			maxBg := ebiten.NewImage(50, 22)
			maxBg.Fill(color.RGBA{255, 215, 0, 255})
			maxBgOp := &ebiten.DrawImageOptions{}
			maxBgOp.GeoM.Translate(485, float64(y+20))
			screen.DrawImage(maxBg, maxBgOp)
			drawText(screen, "MAX", assets.FontSmall, 491, y+25, color.RGBA{0, 0, 0, 255})
		} else {
			canAfford := s.coins >= item.NextCost
			costColor := color.RGBA{255, 100, 100, 255}
			if canAfford {
				costColor = color.RGBA{255, 255, 255, 255}
			}

			costCoinOp := &ebiten.DrawImageOptions{}
			costCoinOp.GeoM.Scale(0.5, 0.5)
			costCoinOp.GeoM.Translate(465, float64(y+17))
			screen.DrawImage(assets.CoinSprite, costCoinOp)

			costText := fmt.Sprintf("%d", item.NextCost)
			drawText(screen, costText, assets.FontSmall, 490, y+32, costColor)
		}
	}

	if s.showUpgradeMsg {
		msgBg := ebiten.NewImage(180, 35)
		msgBg.Fill(colorUpgradeSuccess)
		msgBgOp := &ebiten.DrawImageOptions{}
		msgBgOp.GeoM.Translate(float64((config.ScreenWidth-180)/2), float64(config.ScreenHeight-75))
		screen.DrawImage(msgBg, msgBgOp)

		msgText := "UPGRADED!"
		msgX := (config.ScreenWidth - measureText(msgText, assets.FontSmall)) / 2
		drawText(screen, msgText, assets.FontSmall, msgX, config.ScreenHeight-60, color.White)
	}

	hintBg := ebiten.NewImage(config.ScreenWidth, 25)
	hintBg.Fill(colorHintBackground)
	hintBgOp := &ebiten.DrawImageOptions{}
	hintBgOp.GeoM.Translate(0, float64(config.ScreenHeight-32))
	screen.DrawImage(hintBg, hintBgOp)

	hintText := "ENTER: Buy  |  ESC: Close"
	hintX := (config.ScreenWidth - measureText(hintText, assets.FontSmall)) / 2
	drawText(screen, hintText, assets.FontSmall, hintX, config.ScreenHeight-22, color.RGBA{200, 200, 200, 255})
}

func (s *Shop) GetUpgradeType() string {
	return s.upgradeType
}

func (s *Shop) Reset() {
	s.action = ShopActionNone
	s.selectedIndex = 0
	s.showUpgradeMsg = false
	s.msgTimer = 0
}
