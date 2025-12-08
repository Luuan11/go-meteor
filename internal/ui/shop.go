package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
)

const (
	shopDescriptionOffsetY = 38
	shopPreviewOffsetX     = 320
	shopPreviewOffsetY     = 45
	shopMaxTextOffsetX     = 230
)

var (
	colorUpgradePreview = color.RGBA{100, 255, 150, 255}
	colorMaxLevel       = color.RGBA{255, 215, 0, 255}
	colorDescription    = color.RGBA{255, 200, 100, 255}
	colorSelectedItemBg = color.RGBA{60, 60, 100, 230}
	colorUpgradeSuccess = color.RGBA{0, 200, 0, 230}
	colorHintBackground = color.RGBA{20, 20, 40, 230}

	shopOverlay      *ebiten.Image
	shopHintBg       *ebiten.Image
	dialogOverlay    *ebiten.Image
	upgradeSuccessBg *ebiten.Image
)

func init() {
	shopOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	shopOverlay.Fill(color.RGBA{0, 0, 0, 220})

	shopHintBg = ebiten.NewImage(config.ScreenWidth, 25)
	shopHintBg.Fill(colorHintBackground)

	dialogOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	dialogOverlay.Fill(color.RGBA{0, 0, 0, 180})

	upgradeSuccessBg = ebiten.NewImage(180, 35)
	upgradeSuccessBg.Fill(colorUpgradeSuccess)
}

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
	Items             []ShopItem
	selectedIndex     int
	scrollOffset      int
	maxVisibleItems   int
	coins             int
	action            ShopAction
	upgradeType       string
	showUpgradeMsg    bool
	msgTimer          int
	isMobile          bool
	showConfirmation  bool
	confirmationItem  int
	confirmationCost  int
	confirmationPower string
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

func (s *Shop) SetMobile(isMobile bool) {
	s.isMobile = isMobile
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

	if s.showConfirmation {
		handleConfirmClick := func(clickX, clickY int) {
			if clickX >= config.ScreenWidth/2-110 && clickX <= config.ScreenWidth/2-10 &&
				clickY >= config.ScreenHeight/2+20 && clickY <= config.ScreenHeight/2+60 {
				s.action = ShopActionUpgrade
				s.upgradeType = s.confirmationPower
				s.showUpgradeMsg = true
				s.msgTimer = 60
				s.showConfirmation = false
				return
			}
			// Cancel button
			if clickX >= config.ScreenWidth/2+10 && clickX <= config.ScreenWidth/2+110 &&
				clickY >= config.ScreenHeight/2+20 && clickY <= config.ScreenHeight/2+60 {
				s.showConfirmation = false
				return
			}
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			handleConfirmClick(mouseX, mouseY)
		}

		touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
		if len(touchIDs) > 0 {
			touchX, touchY := ebiten.TouchPosition(touchIDs[0])
			handleConfirmClick(touchX, touchY)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.showConfirmation = false
		}

		return s.action
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
				s.showConfirmation = true
				s.confirmationItem = s.selectedIndex
				s.confirmationCost = item.NextCost
				s.confirmationPower = item.PowerType
			}
		}
	}

	handleClick := func(clickX, clickY int) {
		if clickX >= 10 && clickX <= 50 && clickY >= 10 && clickY <= 50 {
			s.action = ShopActionClose
			return
		}

		startY := 95
		itemHeight := 62

		endIndex := s.scrollOffset + s.maxVisibleItems
		if endIndex > len(s.Items) {
			endIndex = len(s.Items)
		}

		for i := s.scrollOffset; i < endIndex; i++ {
			displayIndex := i - s.scrollOffset
			y := startY + displayIndex*itemHeight

			itemWidth := int(float64(config.ScreenWidth) * 0.8)
			itemStartX := (config.ScreenWidth - itemWidth) / 2

			if clickX >= itemStartX && clickX <= itemStartX+itemWidth &&
				clickY >= y && clickY <= y+itemHeight-4 {
				s.selectedIndex = i
				item := s.Items[i]
				if item.Level < item.MaxLevel && s.coins >= item.NextCost {
					s.showConfirmation = true
					s.confirmationItem = i
					s.confirmationCost = item.NextCost
					s.confirmationPower = item.PowerType
					return
				}
				break
			}
		}

		if len(s.Items) > s.maxVisibleItems {
			if clickX >= config.ScreenWidth-50 && clickX <= config.ScreenWidth-10 &&
				clickY >= 100 && clickY <= 140 {
				if s.scrollOffset > 0 {
					s.scrollOffset--
				}
			}
			if clickX >= config.ScreenWidth-50 && clickX <= config.ScreenWidth-10 &&
				clickY >= config.ScreenHeight-140 && clickY <= config.ScreenHeight-100 {
				if s.scrollOffset < len(s.Items)-s.maxVisibleItems {
					s.scrollOffset++
				}
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		handleClick(mouseX, mouseY)
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		touchX, touchY := ebiten.TouchPosition(touchIDs[0])
		handleClick(touchX, touchY)
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
	screen.DrawImage(shopOverlay, nil)

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

		itemWidth := int(float64(config.ScreenWidth) * 0.8)
		itemStartX := (config.ScreenWidth - itemWidth) / 2

		bgColor := color.RGBA{30, 30, 50, 200}
		borderColor := color.RGBA{60, 60, 90, 255}
		if i == s.selectedIndex {
			bgColor = colorSelectedItemBg
			borderColor = color.RGBA{120, 120, 180, 255}
		}

		border := ebiten.NewImage(itemWidth, itemHeight-4)
		border.Fill(borderColor)
		borderOp := &ebiten.DrawImageOptions{}
		borderOp.GeoM.Translate(float64(itemStartX), float64(y))
		screen.DrawImage(border, borderOp)

		bg := ebiten.NewImage(itemWidth-6, itemHeight-10)
		bg.Fill(bgColor)
		bgOp := &ebiten.DrawImageOptions{}
		bgOp.GeoM.Translate(float64(itemStartX+3), float64(y+3))
		screen.DrawImage(bg, bgOp)

		iconOp := &ebiten.DrawImageOptions{}
		iconOp.GeoM.Scale(0.65, 0.65)
		iconOp.GeoM.Translate(float64(itemStartX+10), float64(y+12))
		screen.DrawImage(item.Icon, iconOp)

		nameText := item.Name
		drawText(screen, nameText, assets.FontSmall, itemStartX+60, y+20, color.White)

		if item.IsSpecial {
			if item.Description != "" {
				drawText(screen, item.Description, assets.FontSmall, itemStartX+60, y+shopDescriptionOffsetY, colorDescription)
			}
		} else {
			levelText := fmt.Sprintf("Level %d/%d", item.Level, item.MaxLevel)
			drawText(screen, levelText, assets.FontSmall, itemStartX+60, y+45, color.RGBA{180, 180, 220, 255})

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
			costCoinOp.GeoM.Translate(610, float64(y+15))
			screen.DrawImage(assets.CoinSprite, costCoinOp)

			costText := fmt.Sprintf("%d", item.NextCost)
			drawText(screen, costText, assets.FontSmall, 630, y+30, costColor)
		}
	}

	if s.showUpgradeMsg {
		msgBgOp := &ebiten.DrawImageOptions{}
		msgBgOp.GeoM.Translate(float64((config.ScreenWidth-180)/2), float64(config.ScreenHeight-75))
		screen.DrawImage(upgradeSuccessBg, msgBgOp)

		msgText := "UPGRADED!"
		msgX := (config.ScreenWidth - measureText(msgText, assets.FontSmall)) / 2
		drawText(screen, msgText, assets.FontSmall, msgX, config.ScreenHeight-60, color.White)
	}

	hintBgOp := &ebiten.DrawImageOptions{}
	hintBgOp.GeoM.Translate(0, float64(config.ScreenHeight-32))
	screen.DrawImage(shopHintBg, hintBgOp)

	hintText := "ENTER: Buy  |  ESC: Close  "
	if s.isMobile {
		hintText = "TOUCH TO BUY"
	}
	hintX := (config.ScreenWidth - measureText(hintText, assets.FontSmall)) / 2
	drawText(screen, hintText, assets.FontSmall, hintX, config.ScreenHeight-22, color.RGBA{200, 200, 200, 255})

	s.drawBackButton(screen)

	if len(s.Items) > s.maxVisibleItems {
		s.drawScrollButtons(screen)
	}

	if s.showConfirmation {
		s.drawConfirmationDialog(screen)
	}
}

func (s *Shop) drawConfirmationDialog(screen *ebiten.Image) {
	// Semi-transparent overlay
	screen.DrawImage(dialogOverlay, nil)

	// Dialog box
	dialogW, dialogH := 400, 200
	dialogX := (config.ScreenWidth - dialogW) / 2
	dialogY := (config.ScreenHeight - dialogH) / 2

	dialog := ebiten.NewImage(dialogW, dialogH)
	dialog.Fill(color.RGBA{30, 30, 50, 255})
	dialogOp := &ebiten.DrawImageOptions{}
	dialogOp.GeoM.Translate(float64(dialogX), float64(dialogY))
	screen.DrawImage(dialog, dialogOp)

	// Border
	border := ebiten.NewImage(dialogW-4, dialogH-4)
	border.Fill(color.RGBA{100, 100, 150, 255})
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(float64(dialogX+2), float64(dialogY+2))
	screen.DrawImage(border, borderOp)

	innerDialog := ebiten.NewImage(dialogW-8, dialogH-8)
	innerDialog.Fill(color.RGBA{30, 30, 50, 255})
	innerDialogOp := &ebiten.DrawImageOptions{}
	innerDialogOp.GeoM.Translate(float64(dialogX+4), float64(dialogY+4))
	screen.DrawImage(innerDialog, innerDialogOp)

	// Title
	titleText := "CONFIRM PURCHASE"
	titleX := dialogX + (dialogW-measureText(titleText, assets.FontSmall))/2
	drawText(screen, titleText, assets.FontSmall, titleX, dialogY+30, color.RGBA{255, 215, 0, 255})

	// Item info
	if s.confirmationItem >= 0 && s.confirmationItem < len(s.Items) {
		item := s.Items[s.confirmationItem]
		itemText := fmt.Sprintf("%s (Lv %d -> %d)", item.Name, item.Level, item.Level+1)
		itemX := dialogX + (dialogW-measureText(itemText, assets.FontSmall))/2
		drawText(screen, itemText, assets.FontSmall, itemX, dialogY+70, color.White)

		costText := fmt.Sprintf("Cost: %d coins", s.confirmationCost)
		costX := dialogX + (dialogW-measureText(costText, assets.FontSmall))/2
		drawText(screen, costText, assets.FontSmall, costX, dialogY+100, color.RGBA{255, 215, 0, 255})
	}

	// Buttons
	btnW, btnH := 100, 40
	confirmX := dialogX + dialogW/2 - btnW - 10
	cancelX := dialogX + dialogW/2 + 10
	btnY := dialogY + dialogH - 60

	// Confirm button
	confirmBtn := ebiten.NewImage(btnW, btnH)
	confirmBtn.Fill(color.RGBA{50, 200, 50, 220})
	confirmBtnOp := &ebiten.DrawImageOptions{}
	confirmBtnOp.GeoM.Translate(float64(confirmX), float64(btnY))
	screen.DrawImage(confirmBtn, confirmBtnOp)
	confirmText := "YES"
	confirmTextX := confirmX + (btnW-measureText(confirmText, assets.FontSmall))/2
	drawText(screen, confirmText, assets.FontSmall, confirmTextX, btnY+25, color.White)

	// Cancel button
	cancelBtn := ebiten.NewImage(btnW, btnH)
	cancelBtn.Fill(color.RGBA{200, 50, 50, 220})
	cancelBtnOp := &ebiten.DrawImageOptions{}
	cancelBtnOp.GeoM.Translate(float64(cancelX), float64(btnY))
	screen.DrawImage(cancelBtn, cancelBtnOp)
	cancelText := "NO"
	cancelTextX := cancelX + (btnW-measureText(cancelText, assets.FontSmall))/2
	drawText(screen, cancelText, assets.FontSmall, cancelTextX, btnY+25, color.White)
}

func (s *Shop) drawBackButton(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()
	isHovered := mouseX >= 10 && mouseX <= 50 && mouseY >= 10 && mouseY <= 50

	btnColor := color.RGBA{100, 100, 150, 180}
	if isHovered {
		btnColor = color.RGBA{150, 150, 200, 220}
	}

	backBtn := ebiten.NewImage(40, 40)
	backBtn.Fill(btnColor)
	backBtnOp := &ebiten.DrawImageOptions{}
	backBtnOp.GeoM.Translate(10, 10)
	screen.DrawImage(backBtn, backBtnOp)

	arrowOp := &ebiten.DrawImageOptions{}
	w, h := assets.ScrollArrow.Bounds().Dx(), assets.ScrollArrow.Bounds().Dy()
	// Rotate 180 degrees to point left (arrow points right by default)
	arrowOp.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	arrowOp.GeoM.Rotate(math.Pi)
	arrowOp.GeoM.Translate(30, 30)
	screen.DrawImage(assets.ScrollArrow, arrowOp)
}

func (s *Shop) drawScrollButtons(screen *ebiten.Image) {
	mouseX, mouseY := ebiten.CursorPosition()

	// Scroll up button
	upHovered := mouseX >= config.ScreenWidth-50 && mouseX <= config.ScreenWidth-10 &&
		mouseY >= 100 && mouseY <= 140
	upColor := color.RGBA{100, 100, 150, 180}
	if upHovered && s.scrollOffset > 0 {
		upColor = color.RGBA{150, 150, 200, 220}
	} else if s.scrollOffset == 0 {
		upColor = color.RGBA{60, 60, 80, 120}
	}

	upBtn := ebiten.NewImage(40, 40)
	upBtn.Fill(upColor)
	upBtnOp := &ebiten.DrawImageOptions{}
	upBtnOp.GeoM.Translate(float64(config.ScreenWidth-50), 100)
	screen.DrawImage(upBtn, upBtnOp)

	// Draw arrow sprite pointing up
	arrowOp := &ebiten.DrawImageOptions{}
	w, h := assets.ScrollArrow.Bounds().Dx(), assets.ScrollArrow.Bounds().Dy()
	// Rotate -90 degrees to point up (arrow points right by default)
	arrowOp.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	arrowOp.GeoM.Rotate(-math.Pi / 2)
	arrowOp.GeoM.Translate(float64(config.ScreenWidth-30), 120)
	screen.DrawImage(assets.ScrollArrow, arrowOp)

	// Scroll down button
	downHovered := mouseX >= config.ScreenWidth-50 && mouseX <= config.ScreenWidth-10 &&
		mouseY >= config.ScreenHeight-140 && mouseY <= config.ScreenHeight-100
	downColor := color.RGBA{100, 100, 150, 180}
	if downHovered && s.scrollOffset < len(s.Items)-s.maxVisibleItems {
		downColor = color.RGBA{150, 150, 200, 220}
	} else if s.scrollOffset >= len(s.Items)-s.maxVisibleItems {
		downColor = color.RGBA{60, 60, 80, 120}
	}

	downBtn := ebiten.NewImage(40, 40)
	downBtn.Fill(downColor)
	downBtnOp := &ebiten.DrawImageOptions{}
	downBtnOp.GeoM.Translate(float64(config.ScreenWidth-50), float64(config.ScreenHeight-140))
	screen.DrawImage(downBtn, downBtnOp)

	// Draw arrow sprite pointing down
	arrowDownOp := &ebiten.DrawImageOptions{}
	w2, h2 := assets.ScrollArrow.Bounds().Dx(), assets.ScrollArrow.Bounds().Dy()
	// Rotate 90 degrees to point down (arrow points right by default)
	arrowDownOp.GeoM.Translate(-float64(w2)/2, -float64(h2)/2)
	arrowDownOp.GeoM.Rotate(math.Pi / 2)
	arrowDownOp.GeoM.Translate(float64(config.ScreenWidth-30), float64(config.ScreenHeight-120))
	screen.DrawImage(assets.ScrollArrow, arrowDownOp)
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
