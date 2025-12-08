package core

import (
	"image/color"
	"time"

	"go-meteor/internal/config"
	"go-meteor/internal/effects"
	"go-meteor/internal/entities"
	"go-meteor/internal/input"
	"go-meteor/internal/systems"
	"go-meteor/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

var emptyImage = ebiten.NewImage(1, 1)

func init() {
	emptyImage.Fill(color.White)
}

type Game struct {
	state            config.GameState
	stateBeforePause config.GameState

	meteoSpawnTimer   *systems.Timer
	starSpawnTimer    *systems.Timer
	powerUpSpawnTimer *systems.Timer
	superPowerTimer   *systems.Timer
	comboTimer        *systems.Timer
	menu              *ui.Menu
	pauseMenu         *ui.PauseMenu
	settingsMenu      *ui.Settings
	notification      *ui.Notification
	shop              *ui.Shop

	player           *entities.Player
	meteors          []*entities.Meteor
	stars            []*entities.Star
	lasers           []*entities.Laser
	powerUps         []*entities.PowerUp
	particles        []*effects.Particle
	superPowerActive bool
	slowMotionActive bool
	slowMotionTimer  *systems.Timer
	laserBeamActive  bool
	laserBeamTimer   *systems.Timer
	nukeActive       bool
	nukeTimer        *systems.Timer
	multiplierActive bool
	multiplierTimer  *systems.Timer

	boss                       *entities.Boss
	bossProjectiles            []*entities.BossProjectile
	bossBar                    *ui.BossBar
	bossWarningShown           bool
	bossAnnouncementTimer      int
	bossCooldownTimer          *systems.Timer
	bossDefeated               bool
	bossCount                  int
	bossNoDamage               bool
	postBossInvincibilityTimer *systems.Timer
	isPostBossInvincible       bool

	meteorPool         *entities.MeteorPool
	laserPool          *entities.LaserPool
	powerUpPool        *entities.PowerUpPool
	bossProjectilePool *entities.BossProjectilePool

	coins []*entities.Coin

	score int
	combo int
	wave  int

	screenShake           int
	playerDeathTimer      int
	playerDeathExplosionX float64
	playerDeathExplosionY float64

	joystick      *input.Joystick
	shootButton   *input.ShootButton
	isMobile      bool
	touchDetected bool

	pauseIconX int
	pauseIconY int

	leaderboard *systems.Leaderboard
	storage     systems.Storage
	highScore   int
	lastScore   int
	progress    *systems.PlayerProgress

	// Game Statistics
	meteorsDestroyed  int
	powerUpsCollected int
	gameStartTime     time.Time
	survivalTime      time.Duration
	statistics        *ui.Statistics
}

func NewGame() *Game {
	g := &Game{
		state:                      config.StateMenu,
		meteoSpawnTimer:            systems.NewTimer(config.MeteorSpawnTime),
		starSpawnTimer:             systems.NewTimer(config.StarSpawnTime),
		powerUpSpawnTimer:          systems.NewTimer(config.PowerUpSpawnTime),
		superPowerTimer:            systems.NewTimer(config.SuperPowerTime),
		comboTimer:                 systems.NewTimer(config.ComboTimeout),
		bossCooldownTimer:          systems.NewTimer(config.BossCooldownTime),
		postBossInvincibilityTimer: systems.NewTimer(config.PostBossInvincibilityTime),
		isPostBossInvincible:       false,
		superPowerActive:           false,
		slowMotionActive:           false,
		slowMotionTimer:            systems.NewTimer(config.SlowMotionTime),
		laserBeamActive:            false,
		laserBeamTimer:             systems.NewTimer(config.LaserBeamTime),
		nukeActive:                 false,
		nukeTimer:                  systems.NewTimer(config.NukeClearScreenTime),
		multiplierActive:           false,
		multiplierTimer:            systems.NewTimer(config.MultiplierTime),
		meteorPool:                 entities.NewMeteorPool(),
		laserPool:                  entities.NewLaserPool(),
		powerUpPool:                entities.NewPowerUpPool(),
		bossProjectilePool:         entities.NewBossProjectilePool(),
		notification:               ui.NewNotification(),
		wave:                       1,
		isMobile:                   false,
		touchDetected:              false,
		leaderboard:                systems.NewLeaderboard(),
		storage:                    systems.NewStorage(),
		meteors:                    make([]*entities.Meteor, 0, 50),
		stars:                      make([]*entities.Star, 0, 50),
		lasers:                     make([]*entities.Laser, 0, config.InitialCapacityLasers),
		powerUps:                   make([]*entities.PowerUp, 0, 10),
		particles:                  make([]*effects.Particle, 0, config.InitialCapacityParticles),
		bossProjectiles:            make([]*entities.BossProjectile, 0, 20),
		coins:                      make([]*entities.Coin, 0, 20),
		bossBar:                    ui.NewBossBar(),
		bossWarningShown:           false,
		bossAnnouncementTimer:      0,
		pauseIconX:                 config.ScreenWidth - (config.PauseIconSize + config.PauseIconMargin),
		pauseIconY:                 config.PauseIconMargin,
	}

	g.player = entities.NewPlayer(g)
	g.menu = ui.NewMenu()
	g.pauseMenu = ui.NewPauseMenu()
	g.settingsMenu = ui.NewSettings()
	g.shop = ui.NewShop()
	g.initMobileControls()

	g.loadHighScore()
	g.loadLeaderboard()
	g.loadProgress()

	return g
}
