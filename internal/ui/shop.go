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
	shopItemHeight      = 62
	shopSeparatorHeight = 60
	shopStartY          = 95
	shopItemWidth       = 0.8

	shopIconOffsetX     = 10
	shopIconOffsetY     = 12
	shopIconScale       = 0.65
	shopSkinIconScale   = 0.5
	shopSkinIconOffsetY = 8

	shopNameOffsetX   = 60
	shopNameOffsetY   = 20
	shopStatusOffsetY = 45

	shopPreviewOffsetX = 320
	shopPreviewOffsetY = 45
	shopMaxTextOffsetX = 230

	shopCostCoinOffsetX = 610
	shopCostCoinOffsetY = 15
	shopCostTextOffsetX = 630
	shopCostTextOffsetY = 30
)

var (
	colorGold           = color.RGBA{255, 215, 0, 255}
	colorGreen          = color.RGBA{0, 255, 100, 255}
	colorRed            = color.RGBA{255, 100, 100, 255}
	colorWhite          = color.RGBA{255, 255, 255, 255}
	colorGray           = color.RGBA{180, 180, 180, 255}
	colorLightGray      = color.RGBA{180, 180, 220, 255}
	colorLightBlue      = color.RGBA{100, 200, 255, 255}
	colorUpgradePreview = color.RGBA{100, 255, 150, 255}
	colorDescription    = color.RGBA{255, 200, 100, 255}

	colorItemBg         = color.RGBA{30, 30, 50, 200}
	colorItemBorder     = color.RGBA{60, 60, 90, 255}
	colorSelectedBg     = color.RGBA{60, 60, 100, 230}
	colorSelectedBorder = color.RGBA{120, 120, 180, 255}

	colorBtnNormal   = color.RGBA{100, 100, 150, 180}
	colorBtnHover    = color.RGBA{150, 150, 200, 220}
	colorBtnDisabled = color.RGBA{60, 60, 80, 120}

	shopOverlay      *ebiten.Image
	dialogOverlay    *ebiten.Image
	upgradeSuccessBg *ebiten.Image
)

func init() {
	shopOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	shopOverlay.Fill(color.RGBA{0, 0, 0, 220})

	dialogOverlay = ebiten.NewImage(config.ScreenWidth, config.ScreenHeight)
	dialogOverlay.Fill(color.RGBA{0, 0, 0, 180})

	upgradeSuccessBg = ebiten.NewImage(180, 35)
	upgradeSuccessBg.Fill(color.RGBA{0, 200, 0, 230})
}

type ShopAction int

const (
	ShopActionNone ShopAction = iota
	ShopActionClose
	ShopActionUpgrade
	ShopActionBuySkin
	ShopActionEquipSkin
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

type SkinItem struct {
	ID         string
	Name       string
	Icon       *ebiten.Image
	Cost       int
	IsOwned    bool
	IsEquipped bool
}

type Shop struct {
	Items            []ShopItem
	Skins            []SkinItem
	AllItems         []interface{}
	selectedIndex    int
	scrollOffset     int
	maxVisibleItems  int
	coins            int
	action           ShopAction
	upgradeType      string
	skinID           string
	showUpgradeMsg   bool
	msgTimer         int
	isMobile         bool
	showConfirmation bool
	confirmCost      int
	confirmPower     string
	confirmType      string
}

func NewShop() *Shop {
	return &Shop{
		Items:           make([]ShopItem, 0),
		Skins:           make([]SkinItem, 0),
		AllItems:        make([]interface{}, 0),
		maxVisibleItems: 7,
		action:          ShopActionNone,
	}
}

func (s *Shop) SetProgress(progress *systems.PlayerProgress) {
	if progress == nil {
		return
	}

	s.coins = progress.Coins
	s.loadShopItems(progress)
	s.loadSkins(progress)
	s.buildAllItems()
	s.scrollOffset = 0
	s.selectedIndex = 0
}

func (s *Shop) loadShopItems(progress *systems.PlayerProgress) {
	s.Items = []ShopItem{
		{PowerType: "superpower", Name: "Super Shot", Icon: assets.SuperPowerSprite, Level: progress.GetUpgradeLevel("superpower"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "shield", Name: "Shield", Icon: assets.ShieldPowerUpSprite, Level: progress.GetUpgradeLevel("shield"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "slowmotion", Name: "Slow Motion", Icon: assets.ClockPowerUpSprite, Level: progress.GetUpgradeLevel("slowmotion"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "laser", Name: "Laser Beam", Icon: assets.LaserPowerUpSprite, Level: progress.GetUpgradeLevel("laser"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "nuke", Name: "Nuke", Icon: assets.NukePowerUpSprite, Level: progress.GetUpgradeLevel("nuke"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "multiplier", Name: "Multiplier", Icon: assets.MultiplierPowerUpSprite, Level: progress.GetUpgradeLevel("multiplier"), MaxLevel: 5, BonusPerLvl: 2},
		{PowerType: "coinmagnet", Name: "Coin Magnet", Icon: assets.CoinSprite, Level: progress.GetUpgradeLevel("coinmagnet"), MaxLevel: 1, IsSpecial: true},
		{PowerType: "doublecoins", Name: "Double Coins", Icon: assets.CoinSprite, Level: progress.GetUpgradeLevel("doublecoins"), MaxLevel: 1, IsSpecial: true},
		{PowerType: "startboost", Name: "Start with Boost", Icon: assets.SuperPowerSprite, Level: progress.GetUpgradeLevel("startboost"), MaxLevel: 1, IsSpecial: true},
	}

	for i := range s.Items {
		s.Items[i].NextCost = s.calculateItemCost(&s.Items[i])
	}
}

func (s *Shop) loadSkins(progress *systems.PlayerProgress) {
	skins := []struct {
		id   string
		name string
		icon *ebiten.Image
		cost int
	}{
		{"gray", "Gray", assets.SkinGray, 0},
		{"green", "Green", assets.SkinGreen, 50},
		{"yellow", "Yellow", assets.SkinYellow, 50},
		{"pink", "Pink", assets.SkinPink, 50},
		{"red", "Red", assets.SkinRed, 50},
		{"purple", "Purple", assets.SkinPurple, 50},
		{"black", "Black", assets.SkinBlack, 100},
		{"gold", "Gold", assets.SkinGold, 250},
		{"white", "White", assets.SkinWhite, 50},
	}

	s.Skins = make([]SkinItem, 0, len(skins))
	for _, skin := range skins {
		s.Skins = append(s.Skins, SkinItem{
			ID:         skin.id,
			Name:       skin.name,
			Icon:       skin.icon,
			Cost:       skin.cost,
			IsOwned:    progress.HasSkin(skin.id),
			IsEquipped: progress.EquippedSkin == skin.id,
		})
	}
}

func (s *Shop) buildAllItems() {
	s.AllItems = make([]interface{}, 0, len(s.Items)+len(s.Skins)+1)
	for i := range s.Items {
		s.AllItems = append(s.AllItems, &s.Items[i])
	}

	s.AllItems = append(s.AllItems, "----- SKINS -----")
	for i := range s.Skins {
		s.AllItems = append(s.AllItems, &s.Skins[i])
	}
}

func (s *Shop) calculateItemCost(item *ShopItem) int {
	if item.Level >= item.MaxLevel {
		return 0
	}

	if item.IsSpecial {
		specialCosts := map[string]int{
			"coinmagnet":  500,
			"doublecoins": 250,
			"startboost":  250,
		}
		if cost, ok := specialCosts[item.PowerType]; ok {
			return cost
		}
		return 0
	}

	costs := []int{25, 50, 100, 200, 400}
	if item.Level < len(costs) {
		return costs[item.Level]
	}
	return 0
}

func (s *Shop) SetMobile(isMobile bool) {
	s.isMobile = isMobile
}

func (s *Shop) Update() ShopAction {
	s.action = ShopActionNone
	s.updateMessageTimer()

	if s.showConfirmation {
		return s.updateConfirmDialog()
	}

	s.updateKeyboardInput()
	s.updateMouseAndTouch()
	return s.action
}

func (s *Shop) updateMessageTimer() {
	if s.msgTimer > 0 {
		s.msgTimer--
		if s.msgTimer == 0 {
			s.showUpgradeMsg = false
		}
	}
}

func (s *Shop) updateConfirmDialog() ShopAction {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.showConfirmation = false
		return s.action
	}

	handleClick := func(x, y int) {
		if s.isConfirmButtonClicked(x, y) {
			s.confirmPurchase()
		} else if s.isCancelButtonClicked(x, y) {
			s.showConfirmation = false
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleClick(ebiten.CursorPosition())
	}

	if touchIDs := inpututil.AppendJustPressedTouchIDs(nil); len(touchIDs) > 0 {
		handleClick(ebiten.TouchPosition(touchIDs[0]))
	}

	return s.action
}

func (s *Shop) isConfirmButtonClicked(x, y int) bool {
	btnY := config.ScreenHeight/2 + 20
	return x >= config.ScreenWidth/2-110 && x <= config.ScreenWidth/2-10 &&
		y >= btnY && y <= btnY+40
}

func (s *Shop) isCancelButtonClicked(x, y int) bool {
	btnY := config.ScreenHeight/2 + 20
	return x >= config.ScreenWidth/2+10 && x <= config.ScreenWidth/2+110 &&
		y >= btnY && y <= btnY+40
}

func (s *Shop) confirmPurchase() {
	switch s.confirmType {
	case "upgrade":
		s.action = ShopActionUpgrade
		s.upgradeType = s.confirmPower
	case "buyskin":
		s.action = ShopActionBuySkin
		s.skinID = s.confirmPower
	case "equip":
		s.action = ShopActionEquipSkin
		s.skinID = s.confirmPower
	}
	s.showUpgradeMsg = true
	s.msgTimer = 60
	s.showConfirmation = false
}

func (s *Shop) updateKeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.action = ShopActionClose
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.moveSelectionUp()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.moveSelectionDown()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.handleItemSelection()
	}
}

func (s *Shop) moveSelectionUp() {
	s.selectedIndex--
	if s.selectedIndex < 0 {
		s.selectedIndex = len(s.AllItems) - 1
	}
	s.skipSeparatorBackward()
	s.updateScroll()
}

func (s *Shop) moveSelectionDown() {
	s.selectedIndex++
	if s.selectedIndex >= len(s.AllItems) {
		s.selectedIndex = 0
	}
	s.skipSeparatorForward()
	s.updateScroll()
}

func (s *Shop) skipSeparatorForward() {
	if s.selectedIndex < 0 || s.selectedIndex >= len(s.AllItems) {
		return
	}

	if _, isSep := s.AllItems[s.selectedIndex].(string); isSep {
		s.selectedIndex++
		if s.selectedIndex >= len(s.AllItems) {
			s.selectedIndex = 0
		}
	}
}

func (s *Shop) skipSeparatorBackward() {
	if s.selectedIndex < 0 || s.selectedIndex >= len(s.AllItems) {
		return
	}

	if _, isSep := s.AllItems[s.selectedIndex].(string); isSep {
		s.selectedIndex--
		if s.selectedIndex < 0 {
			s.selectedIndex = len(s.AllItems) - 1
		}
	}
}

func (s *Shop) updateMouseAndTouch() {
	handleClick := func(x, y int) {
		if s.isBackButtonClicked(x, y) {
			s.action = ShopActionClose
			return
		}

		if s.handleScrollButtonClick(x, y) {
			return
		}

		s.handleItemClick(x, y)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleClick(ebiten.CursorPosition())
	}

	if touchIDs := inpututil.AppendJustPressedTouchIDs(nil); len(touchIDs) > 0 {
		handleClick(ebiten.TouchPosition(touchIDs[0]))
	}
}

func (s *Shop) isBackButtonClicked(x, y int) bool {
	return x >= 10 && x <= 50 && y >= 10 && y <= 50
}

func (s *Shop) handleScrollButtonClick(x, y int) bool {
	if len(s.AllItems) <= s.maxVisibleItems {
		return false
	}

	if x >= config.ScreenWidth-50 && x <= config.ScreenWidth-10 {
		if y >= 100 && y <= 140 && s.scrollOffset > 0 {
			s.scrollOffset--
			return true
		}
		if y >= config.ScreenHeight-140 && y <= config.ScreenHeight-100 &&
			s.scrollOffset < len(s.AllItems)-s.maxVisibleItems {
			s.scrollOffset++
			return true
		}
	}
	return false
}

func (s *Shop) handleItemClick(x, y int) {
	itemWidth := int(float64(config.ScreenWidth) * shopItemWidth)
	itemStartX := (config.ScreenWidth - itemWidth) / 2
	currentY := shopStartY

	endIndex := s.scrollOffset + s.maxVisibleItems
	if endIndex > len(s.AllItems) {
		endIndex = len(s.AllItems)
	}

	for i := s.scrollOffset; i < endIndex; i++ {
		if _, isSep := s.AllItems[i].(string); isSep {
			currentY += shopSeparatorHeight
			continue
		}

		if x >= itemStartX && x <= itemStartX+itemWidth &&
			y >= currentY && y <= currentY+shopItemHeight-4 {
			s.selectedIndex = i
			s.handleItemSelection()
			return
		}

		currentY += shopItemHeight
	}
}

func (s *Shop) handleItemSelection() {
	if s.selectedIndex < 0 || s.selectedIndex >= len(s.AllItems) {
		return
	}

	switch item := s.AllItems[s.selectedIndex].(type) {
	case *ShopItem:
		s.handleShopItemSelection(item)
	case *SkinItem:
		s.handleSkinItemSelection(item)
	}
}

func (s *Shop) handleShopItemSelection(item *ShopItem) {
	if item.Level < item.MaxLevel && s.coins >= item.NextCost {
		s.showConfirmation = true
		s.confirmCost = item.NextCost
		s.confirmPower = item.PowerType
		s.confirmType = "upgrade"
	}
}

func (s *Shop) handleSkinItemSelection(item *SkinItem) {
	if !item.IsOwned && s.coins >= item.Cost {
		s.showConfirmation = true
		s.confirmCost = item.Cost
		s.confirmPower = item.ID
		s.confirmType = "buyskin"
	} else if item.IsOwned && !item.IsEquipped {
		s.action = ShopActionEquipSkin
		s.skinID = item.ID
	}
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
	s.drawHeader(screen)
	s.drawAllItems(screen)
	s.drawSuccessMessage(screen)
	s.drawHint(screen)
	s.drawBackButton(screen)

	if len(s.AllItems) > s.maxVisibleItems {
		s.drawScrollButtons(screen)
	}

	if s.showConfirmation {
		s.drawConfirmDialog(screen)
	}
}

func (s *Shop) drawHeader(screen *ebiten.Image) {
	titleText := "SHOP"
	titleX := (config.ScreenWidth - measureText(titleText, assets.FontUi)) / 2
	drawText(screen, titleText, assets.FontUi, titleX, 40, colorGold)

	coinOp := &ebiten.DrawImageOptions{}
	coinOp.GeoM.Scale(0.8, 0.8)
	coinOp.GeoM.Translate(float64(config.ScreenWidth/2-25), 55)
	screen.DrawImage(assets.CoinSprite, coinOp)

	coinsText := fmt.Sprintf("%d", s.coins)
	coinsX := (config.ScreenWidth - measureText(coinsText, assets.FontSmall)) / 2
	drawText(screen, coinsText, assets.FontSmall, coinsX+25, 75, colorWhite)
}

func (s *Shop) drawSuccessMessage(screen *ebiten.Image) {
	if !s.showUpgradeMsg {
		return
	}

	msgBgOp := &ebiten.DrawImageOptions{}
	msgBgOp.GeoM.Translate(float64((config.ScreenWidth-180)/2), float64(config.ScreenHeight-75))
	screen.DrawImage(upgradeSuccessBg, msgBgOp)

	msgText := "SUCCESS!"
	msgX := (config.ScreenWidth - measureText(msgText, assets.FontSmall)) / 2
	drawText(screen, msgText, assets.FontSmall, msgX, config.ScreenHeight-60, colorWhite)
}

func (s *Shop) drawHint(screen *ebiten.Image) {
	hintBg := ebiten.NewImage(config.ScreenWidth, 25)
	hintBg.Fill(color.RGBA{20, 20, 40, 230})
	hintBgOp := &ebiten.DrawImageOptions{}
	hintBgOp.GeoM.Translate(0, float64(config.ScreenHeight-32))
	screen.DrawImage(hintBg, hintBgOp)

	hintText := "ENTER: Select  |  ESC: Close"
	if s.isMobile {
		hintText = "TAP TO SELECT"
	}
	hintX := (config.ScreenWidth - measureText(hintText, assets.FontSmall)) / 2
	drawText(screen, hintText, assets.FontSmall, hintX, config.ScreenHeight-22, color.RGBA{200, 200, 200, 255})
}

func (s *Shop) drawBackButton(screen *ebiten.Image) {
	x, y := ebiten.CursorPosition()
	btnColor := colorBtnNormal
	if x >= 10 && x <= 50 && y >= 10 && y <= 50 {
		btnColor = colorBtnHover
	}

	backBtn := ebiten.NewImage(40, 40)
	backBtn.Fill(btnColor)
	backBtnOp := &ebiten.DrawImageOptions{}
	backBtnOp.GeoM.Translate(10, 10)
	screen.DrawImage(backBtn, backBtnOp)

	s.drawArrow(screen, 30, 30, math.Pi)
}

func (s *Shop) drawScrollButtons(screen *ebiten.Image) {
	x, y := ebiten.CursorPosition()
	maxItems := len(s.AllItems)

	upColor := s.getScrollButtonColor(x, y, config.ScreenWidth-50, config.ScreenWidth-10, 100, 140, s.scrollOffset > 0)
	s.drawScrollButton(screen, config.ScreenWidth-50, 100, upColor)
	s.drawArrow(screen, float64(config.ScreenWidth-30), 120, -math.Pi/2)

	downColor := s.getScrollButtonColor(x, y, config.ScreenWidth-50, config.ScreenWidth-10,
		config.ScreenHeight-140, config.ScreenHeight-100, s.scrollOffset < maxItems-s.maxVisibleItems)
	s.drawScrollButton(screen, config.ScreenWidth-50, config.ScreenHeight-140, downColor)
	s.drawArrow(screen, float64(config.ScreenWidth-30), float64(config.ScreenHeight-120), math.Pi/2)
}

func (s *Shop) getScrollButtonColor(mouseX, mouseY, x1, x2, y1, y2 int, canScroll bool) color.Color {
	isHovered := mouseX >= x1 && mouseX <= x2 && mouseY >= y1 && mouseY <= y2
	if !canScroll {
		return colorBtnDisabled
	}
	if isHovered {
		return colorBtnHover
	}
	return colorBtnNormal
}

func (s *Shop) drawScrollButton(screen *ebiten.Image, x, y int, col color.Color) {
	btn := ebiten.NewImage(40, 40)
	btn.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
}

func (s *Shop) drawArrow(screen *ebiten.Image, x, y, rotation float64) {
	op := &ebiten.DrawImageOptions{}
	w, h := assets.ScrollArrow.Bounds().Dx(), assets.ScrollArrow.Bounds().Dy()
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(x, y)
	screen.DrawImage(assets.ScrollArrow, op)
}

func (s *Shop) drawConfirmDialog(screen *ebiten.Image) {
	screen.DrawImage(dialogOverlay, nil)

	dialogW, dialogH := 400, 200
	dialogX := (config.ScreenWidth - dialogW) / 2
	dialogY := (config.ScreenHeight - dialogH) / 2

	s.drawDialogBox(screen, dialogX, dialogY, dialogW, dialogH)
	s.drawDialogContent(screen, dialogX, dialogY, dialogW, dialogH)
	s.drawDialogButtons(screen, dialogX, dialogY, dialogW, dialogH)
}

func (s *Shop) drawDialogBox(screen *ebiten.Image, x, y, w, h int) {
	dialog := ebiten.NewImage(w, h)
	dialog.Fill(color.RGBA{30, 30, 50, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(dialog, op)

	border := ebiten.NewImage(w-4, h-4)
	border.Fill(color.RGBA{100, 100, 150, 255})
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(float64(x+2), float64(y+2))
	screen.DrawImage(border, borderOp)

	inner := ebiten.NewImage(w-8, h-8)
	inner.Fill(color.RGBA{30, 30, 50, 255})
	innerOp := &ebiten.DrawImageOptions{}
	innerOp.GeoM.Translate(float64(x+4), float64(y+4))
	screen.DrawImage(inner, innerOp)
}

func (s *Shop) drawDialogContent(screen *ebiten.Image, dialogX, dialogY, dialogW, dialogH int) {
	titleText := "CONFIRM PURCHASE"
	titleX := dialogX + (dialogW-measureText(titleText, assets.FontSmall))/2
	drawText(screen, titleText, assets.FontSmall, titleX, dialogY+30, colorGold)

	itemText := s.getConfirmItemText()
	itemX := dialogX + (dialogW-measureText(itemText, assets.FontSmall))/2
	drawText(screen, itemText, assets.FontSmall, itemX, dialogY+70, colorWhite)

	costText := fmt.Sprintf("Cost: %d coins", s.confirmCost)
	costX := dialogX + (dialogW-measureText(costText, assets.FontSmall))/2
	drawText(screen, costText, assets.FontSmall, costX, dialogY+100, colorGold)
}

func (s *Shop) getConfirmItemText() string {
	if s.confirmType == "upgrade" {
		for i := range s.Items {
			if s.Items[i].PowerType == s.confirmPower {
				item := s.Items[i]
				return fmt.Sprintf("%s (Lv %d -> %d)", item.Name, item.Level, item.Level+1)
			}
		}
	} else if s.confirmType == "buyskin" {
		for i := range s.Skins {
			if s.Skins[i].ID == s.confirmPower {
				return fmt.Sprintf("%s Skin", s.Skins[i].Name)
			}
		}
	}
	return ""
}

func (s *Shop) drawDialogButtons(screen *ebiten.Image, dialogX, dialogY, dialogW, dialogH int) {
	btnW, btnH := 100, 40
	confirmX := dialogX + dialogW/2 - btnW - 10
	cancelX := dialogX + dialogW/2 + 10
	btnY := dialogY + dialogH - 60

	s.drawButton(screen, confirmX, btnY, btnW, btnH, "YES", color.RGBA{50, 200, 50, 220})
	s.drawButton(screen, cancelX, btnY, btnW, btnH, "NO", color.RGBA{200, 50, 50, 220})
}

func (s *Shop) drawButton(screen *ebiten.Image, x, y, w, h int, text string, col color.Color) {
	btn := ebiten.NewImage(w, h)
	btn.Fill(col)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	textX := x + (w-measureText(text, assets.FontSmall))/2
	drawText(screen, text, assets.FontSmall, textX, y+25, colorWhite)
}

func (s *Shop) drawAllItems(screen *ebiten.Image) {
	endIndex := s.scrollOffset + s.maxVisibleItems
	if endIndex > len(s.AllItems) {
		endIndex = len(s.AllItems)
	}

	currentY := shopStartY
	for i := s.scrollOffset; i < endIndex; i++ {
		switch item := s.AllItems[i].(type) {
		case string:
			s.drawSeparator(screen, item, currentY)
			currentY += shopSeparatorHeight
		case *ShopItem:
			s.drawShopItem(screen, item, i, currentY)
			currentY += shopItemHeight
		case *SkinItem:
			s.drawSkinItem(screen, item, i, currentY)
			currentY += shopItemHeight
		}
	}
}

func (s *Shop) drawSeparator(screen *ebiten.Image, text string, y int) {
	separatorY := y + shopSeparatorHeight/2 - 10
	textX := (config.ScreenWidth - measureText(text, assets.FontSmall)) / 2
	drawText(screen, text, assets.FontSmall, textX, separatorY, colorGold)

	lineY := separatorY + 20
	lineWidth := 300
	lineX := (config.ScreenWidth - lineWidth) / 2
	line := ebiten.NewImage(lineWidth, 2)
	line.Fill(color.RGBA{255, 215, 0, 180})
	lineOp := &ebiten.DrawImageOptions{}
	lineOp.GeoM.Translate(float64(lineX), float64(lineY))
	screen.DrawImage(line, lineOp)
}

func (s *Shop) drawShopItem(screen *ebiten.Image, item *ShopItem, index, y int) {
	itemWidth := int(float64(config.ScreenWidth) * shopItemWidth)
	itemX := (config.ScreenWidth - itemWidth) / 2

	bgColor, borderColor := s.getItemColors(index, false)
	s.drawItemBox(screen, itemX, y, itemWidth, bgColor, borderColor)

	iconOp := &ebiten.DrawImageOptions{}
	iconOp.GeoM.Scale(shopIconScale, shopIconScale)
	iconOp.GeoM.Translate(float64(itemX+shopIconOffsetX), float64(y+shopIconOffsetY))
	screen.DrawImage(item.Icon, iconOp)

	drawText(screen, item.Name, assets.FontSmall, itemX+shopNameOffsetX, y+shopNameOffsetY, colorWhite)

	if item.IsSpecial {
		if item.Description != "" {
			drawText(screen, item.Description, assets.FontSmall, itemX+shopNameOffsetX, y+38, colorDescription)
		}
	} else {
		s.drawShopItemLevel(screen, item, itemX, y)
	}

	s.drawShopItemCost(screen, item, y)
}

func (s *Shop) drawShopItemLevel(screen *ebiten.Image, item *ShopItem, x, y int) {
	levelText := fmt.Sprintf("Level %d/%d", item.Level, item.MaxLevel)
	drawText(screen, levelText, assets.FontSmall, x+shopNameOffsetX, y+shopStatusOffsetY, colorLightGray)

	if item.Level < item.MaxLevel {
		currentDuration := 10 + (item.Level * item.BonusPerLvl)
		nextDuration := 10 + ((item.Level + 1) * item.BonusPerLvl)
		previewText := fmt.Sprintf("%ds - %ds", currentDuration, nextDuration)
		drawText(screen, previewText, assets.FontSmall, shopPreviewOffsetX, y+shopPreviewOffsetY, colorUpgradePreview)
	} else {
		drawText(screen, "MAX", assets.FontSmall, shopMaxTextOffsetX, y+shopPreviewOffsetY, colorGold)
	}
}

func (s *Shop) drawShopItemCost(screen *ebiten.Image, item *ShopItem, y int) {
	if item.Level >= item.MaxLevel {
		maxBg := ebiten.NewImage(50, 22)
		maxBg.Fill(colorGold)
		maxBgOp := &ebiten.DrawImageOptions{}
		maxBgOp.GeoM.Translate(485, float64(y+20))
		screen.DrawImage(maxBg, maxBgOp)
		drawText(screen, "MAX", assets.FontSmall, 491, y+25, color.RGBA{0, 0, 0, 255})
	} else {
		s.drawCostInfo(screen, item.NextCost, y)
	}
}

func (s *Shop) drawSkinItem(screen *ebiten.Image, item *SkinItem, index, y int) {
	itemWidth := int(float64(config.ScreenWidth) * shopItemWidth)
	itemX := (config.ScreenWidth - itemWidth) / 2

	bgColor, borderColor := s.getItemColors(index, item.IsEquipped)
	s.drawItemBox(screen, itemX, y, itemWidth, bgColor, borderColor)

	iconOp := &ebiten.DrawImageOptions{}
	iconOp.GeoM.Scale(shopSkinIconScale, shopSkinIconScale)
	iconOp.GeoM.Translate(float64(itemX+shopIconOffsetX), float64(y+shopSkinIconOffsetY))
	screen.DrawImage(item.Icon, iconOp)

	drawText(screen, item.Name, assets.FontSmall, itemX+shopNameOffsetX, y+shopNameOffsetY, colorWhite)
	s.drawSkinStatus(screen, item, itemX, y)
	s.drawSkinCostOrCheck(screen, item, y)
}

func (s *Shop) drawSkinStatus(screen *ebiten.Image, item *SkinItem, x, y int) {
	statusText, statusColor := s.getSkinStatusText(item)
	drawText(screen, statusText, assets.FontSmall, x+shopNameOffsetX, y+shopStatusOffsetY, statusColor)
}

func (s *Shop) getSkinStatusText(item *SkinItem) (string, color.Color) {
	if item.IsEquipped {
		return "EQUIPPED", colorGreen
	}
	if item.IsOwned {
		return "Click to equip", colorLightBlue
	}
	return "", colorGray
}

func (s *Shop) drawSkinCostOrCheck(screen *ebiten.Image, item *SkinItem, y int) {
	if item.IsOwned {
		if item.IsEquipped {
			checkBg := ebiten.NewImage(30, 30)
			checkBg.Fill(color.RGBA{0, 200, 0, 255})
			checkBgOp := &ebiten.DrawImageOptions{}
			checkBgOp.GeoM.Translate(640, float64(y+15))
			screen.DrawImage(checkBg, checkBgOp)
		}
	} else {
		s.drawCostInfo(screen, item.Cost, y)
	}
}

func (s *Shop) drawCostInfo(screen *ebiten.Image, cost, y int) {
	costColor := colorRed
	if s.coins >= cost {
		costColor = colorWhite
	}

	coinOp := &ebiten.DrawImageOptions{}
	coinOp.GeoM.Scale(0.5, 0.5)
	coinOp.GeoM.Translate(shopCostCoinOffsetX, float64(y+shopCostCoinOffsetY))
	screen.DrawImage(assets.CoinSprite, coinOp)

	costText := fmt.Sprintf("%d", cost)
	drawText(screen, costText, assets.FontSmall, shopCostTextOffsetX, y+shopCostTextOffsetY, costColor)
}

func (s *Shop) getItemColors(index int, isEquipped bool) (bgColor, borderColor color.Color) {
	bgColor = colorItemBg
	borderColor = colorItemBorder

	if index == s.selectedIndex {
		bgColor = colorSelectedBg
		borderColor = colorSelectedBorder
	}

	if isEquipped {
		borderColor = colorGreen
	}

	return
}

func (s *Shop) drawItemBox(screen *ebiten.Image, x, y, width int, bgColor, borderColor color.Color) {
	border := ebiten.NewImage(width, shopItemHeight-4)
	border.Fill(borderColor)
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, borderOp)

	bg := ebiten.NewImage(width-6, shopItemHeight-10)
	bg.Fill(bgColor)
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(x+3), float64(y+3))
	screen.DrawImage(bg, bgOp)
}

func (s *Shop) GetUpgradeType() string {
	return s.upgradeType
}

func (s *Shop) GetSkinID() string {
	return s.skinID
}

func (s *Shop) Reset() {
	s.action = ShopActionNone
	s.selectedIndex = 0
	s.scrollOffset = 0
	s.showUpgradeMsg = false
	s.msgTimer = 0
}
