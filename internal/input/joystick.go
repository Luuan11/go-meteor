package input

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	touchDebounceDelay = 100
)

type Joystick struct {
	centerX    float64
	centerY    float64
	radius     float64
	knobRadius float64
	touchID    ebiten.TouchID
	isActive   bool
	deltaX     float64
	deltaY     float64
}

func NewJoystick(x, y, radius float64) *Joystick {
	return &Joystick{
		centerX:    x,
		centerY:    y,
		radius:     radius,
		knobRadius: radius * 0.4,
		touchID:    -1,
		isActive:   false,
	}
}

func (j *Joystick) Update(touchIDs []ebiten.TouchID) {
	if !j.isActive {
		for _, id := range touchIDs {
			x, y := ebiten.TouchPosition(id)
			fx, fy := float64(x), float64(y)

			dist := math.Sqrt(math.Pow(fx-j.centerX, 2) + math.Pow(fy-j.centerY, 2))
			if dist <= j.radius {
				j.isActive = true
				j.touchID = id
				break
			}
		}
	}

	if j.isActive {
		found := false
		for _, id := range touchIDs {
			if id == j.touchID {
				found = true
				x, y := ebiten.TouchPosition(id)
				fx, fy := float64(x), float64(y)

				dx := fx - j.centerX
				dy := fy - j.centerY
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist > j.radius {
					angle := math.Atan2(dy, dx)
					dx = math.Cos(angle) * j.radius
					dy = math.Sin(angle) * j.radius
				}

				j.deltaX = dx / j.radius
				j.deltaY = dy / j.radius
				break
			}
		}

		if !found {
			j.isActive = false
			j.touchID = -1
			j.deltaX = 0
			j.deltaY = 0
		}
	}
}

func (j *Joystick) Draw(screen *ebiten.Image) {
	baseColor := color.RGBA{100, 100, 100, 120}
	knobColor := color.RGBA{200, 200, 200, 180}

	if j.isActive {
		baseColor = color.RGBA{150, 150, 150, 150}
		knobColor = color.RGBA{255, 255, 255, 220}
	}

	vector.DrawFilledCircle(screen, float32(j.centerX), float32(j.centerY), float32(j.radius), baseColor, false)

	knobX := j.centerX + j.deltaX*j.radius*0.6
	knobY := j.centerY + j.deltaY*j.radius*0.6
	vector.DrawFilledCircle(screen, float32(knobX), float32(knobY), float32(j.knobRadius), knobColor, false)
}

func (j *Joystick) GetDirection() (float64, float64) {
	return j.deltaX, j.deltaY
}

func (j *Joystick) IsPressed() bool {
	return j.isActive && (math.Abs(j.deltaX) > 0.1 || math.Abs(j.deltaY) > 0.1)
}

type ShootButton struct {
	x             float64
	y             float64
	radius        float64
	touchID       ebiten.TouchID
	isActive      bool
	wasPressed    bool
	lastPressTime int64 // For debouncing
}

func NewShootButton(x, y, radius float64) *ShootButton {
	return &ShootButton{
		x:          x,
		y:          y,
		radius:     radius,
		touchID:    -1,
		isActive:   false,
		wasPressed: false,
	}
}

func (sb *ShootButton) Update(touchIDs []ebiten.TouchID) bool {
	pressed := false
	currentTime := time.Now().UnixMilli()

	if !sb.isActive {
		if currentTime-sb.lastPressTime < touchDebounceDelay {
			return false
		}

		for _, id := range touchIDs {
			x, y := ebiten.TouchPosition(id)
			fx, fy := float64(x), float64(y)

			dist := math.Sqrt(math.Pow(fx-sb.x, 2) + math.Pow(fy-sb.y, 2))
			if dist <= sb.radius {
				sb.isActive = true
				sb.touchID = id
				if !sb.wasPressed {
					pressed = true
					sb.wasPressed = true
					sb.lastPressTime = currentTime
				}
				break
			}
		}
	} else {
		found := false
		for _, id := range touchIDs {
			if id == sb.touchID {
				found = true
				break
			}
		}

		if !found {
			sb.isActive = false
			sb.touchID = -1
			sb.wasPressed = false
		}
	}

	return pressed
}

func (sb *ShootButton) IsActive() bool {
	return sb.isActive
}

func (sb *ShootButton) Draw(screen *ebiten.Image) {
	buttonColor := color.RGBA{255, 100, 100, 120}
	if sb.isActive {
		buttonColor = color.RGBA{255, 150, 150, 180}
	}

	vector.DrawFilledCircle(screen, float32(sb.x), float32(sb.y), float32(sb.radius), buttonColor, false)

	circleColor := color.RGBA{255, 255, 255, 200}
	vector.DrawFilledCircle(screen, float32(sb.x), float32(sb.y), float32(sb.radius*0.4), circleColor, false)
}
