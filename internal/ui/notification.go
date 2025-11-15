package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"

	"go-meteor/internal/config"
	"go-meteor/internal/systems"
	assets "go-meteor/src/pkg"
)

type NotificationType int

const (
	NotificationLife NotificationType = iota
	NotificationShield
	NotificationSuperPower
)

type Notification struct {
	message   string
	timer     *systems.Timer
	notifType NotificationType
	active    bool
}

func NewNotification() *Notification {
	return &Notification{
		timer:  systems.NewTimer(3 * time.Second),
		active: false,
	}
}

func (n *Notification) Show(message string, notifType NotificationType) {
	n.message = message
	n.notifType = notifType
	n.active = true
	n.timer.Reset()
}

func (n *Notification) Update() {
	if !n.active {
		return
	}

	n.timer.Update()
	if n.timer.IsReady() {
		n.active = false
	}
}

func (n *Notification) Draw(screen *ebiten.Image) {
	if !n.active {
		return
	}

	alpha := 1.0
	if n.timer.CurrentTicks() > 120 {
		alpha = 1.0 - float64(n.timer.CurrentTicks()-120)/60.0
		if alpha < 0 {
			alpha = 0
		}
	}

	var textColor color.Color
	switch n.notifType {
	case NotificationLife:
		textColor = color.RGBA{R: 0, G: 255, B: 0, A: uint8(255 * alpha)}
	case NotificationShield:
		textColor = color.RGBA{R: 100, G: 200, B: 255, A: uint8(255 * alpha)}
	case NotificationSuperPower:
		hue := float64(n.timer.CurrentTicks()%60) / 60.0
		r, g, b := hslToRGB(hue, 1.0, 0.6)
		textColor = color.RGBA{R: r, G: g, B: b, A: uint8(255 * alpha)}
	}

	bounds := text.BoundString(assets.FontUi, n.message)
	x := (config.ScreenWidth - bounds.Dx()) / 2
	text.Draw(screen, n.message, assets.FontUi, x, 150, textColor)
}

func (n *Notification) IsActive() bool {
	return n.active
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	var r, g, b float64

	if s == 0 {
		r, g, b = l, l, l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}
