package systems

import (
	"encoding/json"
	"sort"
	"time"
)

const MaxLeaderboardEntries = 10

type LeaderboardEntry struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Date  string `json:"date"`
}

type Leaderboard struct {
	Entries []LeaderboardEntry `json:"entries"`
}

func NewLeaderboard() *Leaderboard {
	return &Leaderboard{
		Entries: make([]LeaderboardEntry, 0, MaxLeaderboardEntries),
	}
}

func (l *Leaderboard) AddScore(name string, score int) {
	entry := LeaderboardEntry{
		Name:  name,
		Score: score,
		Date:  time.Now().Format("2006-01-02"),
	}

	l.Entries = append(l.Entries, entry)
	l.Sort()

	if len(l.Entries) > MaxLeaderboardEntries {
		l.Entries = l.Entries[:MaxLeaderboardEntries]
	}
}

func (l *Leaderboard) IsTopScore(score int) bool {
	if len(l.Entries) < MaxLeaderboardEntries {
		return true
	}
	return score > l.Entries[MaxLeaderboardEntries-1].Score
}

func (l *Leaderboard) Sort() {
	sort.Slice(l.Entries, func(i, j int) bool {
		return l.Entries[i].Score > l.Entries[j].Score
	})
}

func (l *Leaderboard) ToJSON() (string, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (l *Leaderboard) FromJSON(jsonData string) error {
	return json.Unmarshal([]byte(jsonData), l)
}
