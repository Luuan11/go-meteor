# Go-meteor

<img width="1564" height="709" alt="image" src="https://github.com/user-attachments/assets/415f16a6-f511-41fa-9ab0-6ae024d574ed" />

## <p align="center"> <a href="https://luuan11.github.io/go-meteor/">Click here to play this game!</a> </p>


## 💬 About
Interactive game project made entirely with GO, where you can control your player, destroy meteors, collect power-ups, all to increase your score.

## 🎮 Game Features

### Power-Ups System
- **Super Shot**: Triple shot with 2x damage (10s)
- **Health**: Restores 1 life point
- **Shield**: Temporary invincibility (10s)
- **Slow Motion**: Slows down meteors (15s)
- **Laser Beam**: Powerful penetrating shots with 3x damage (3.5s, Wave 5+)
- **Nuke**: Clears screen for 5s while keeping combo (Wave 5+)
- **Extra Life**: Grants 1 extra life, max 5 hearts (Wave 5+)
- **Score Multiplier**: 2x score bonus for 20 seconds

### Boss System
- 3 unique boss types with different behaviors:
  - **Tank**: Slow and heavily armored (150 HP)
  - **Sniper**: Fast and precise attacks (80 HP)
  - **Swarm**: Medium speed with dual shots (100 HP)
- Random boss spawns every 5 waves
- Boss announcement with countdown

### Gameplay Systems
- Combo System with Score Multiplier
- Wave System with Progressive Difficulty (harder from wave 15+)
- Audio System (Background Music and Sound Effects)
- Responsive Controls for Desktop and Mobile
- Global Leaderboard with Top 10 Rankings
- Post-Game Statistics

## 💡 Technical Stack
- Go
- Ebiten package
- Firebase Realtime Database

### Architecture & Design Patterns
- Clean Architecture with separated concerns
- Object Pooling for performance optimization
- State Machine for game flow management
- Component-based entity system

## 📦 Installation

    - Clone repository 
    $ git clone https://github.com/Luuan11/go-meteor.git 

    - Install dependencies
    $ go mod tidy

    - Run application
    $ go run main.go

---

Made with 💜 by [Luan Fernando](https://www.linkedin.com/in/luan-fernando/).
