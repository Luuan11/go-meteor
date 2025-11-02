package entities

import (
	"sync"
)

type MeteorPool struct {
	pool sync.Pool
}

func NewMeteorPool() *MeteorPool {
	return &MeteorPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Meteor{}
			},
		},
	}
}

func (p *MeteorPool) Get() *Meteor {
	return p.pool.Get().(*Meteor)
}

func (p *MeteorPool) Put(m *Meteor) {
	p.pool.Put(m)
}

type LaserPool struct {
	pool sync.Pool
}

func NewLaserPool() *LaserPool {
	return &LaserPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Laser{}
			},
		},
	}
}

func (p *LaserPool) Get() *Laser {
	return p.pool.Get().(*Laser)
}

func (p *LaserPool) Put(l *Laser) {
	p.pool.Put(l)
}

type PowerUpPool struct {
	pool sync.Pool
}

func NewPowerUpPool() *PowerUpPool {
	return &PowerUpPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &PowerUp{}
			},
		},
	}
}

func (p *PowerUpPool) Get() *PowerUp {
	return p.pool.Get().(*PowerUp)
}

func (p *PowerUpPool) Put(pu *PowerUp) {
	p.pool.Put(pu)
}
