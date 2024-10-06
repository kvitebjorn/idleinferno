package model

import (
	"math"
)

type Stats struct {
	Xp      uint64
	Created string
	Online  bool
}

func (s *Stats) IncrementXp() {
	s.Xp += 1
}

func (s *Stats) DecrementXp() {
	if s.Xp > 0 {
		s.Xp -= 1
	}
}

const (
	// With these factors, it takes 229.4 days to reach level 100 @ 1xp per minute.
	// if the player remains logged in 24/7...
	C = 20
	x = 2
)

/*
Base xp formula:
xpNeededToLevelUp=currentLevel^x + C

Therefore, we can express current level as:
xpNeededToLevelUp − C=currentLevel^x
currentLevel=(xp − C)^(1/x)

To find the XP required to reach the next level:
xpNeededToNextLevel=currentLevel^x + C

There is no explicit closed form to calculate the level.
This is due to the sum of series property of the xp formula.
*/

// Level calculates the player's current level based on their XP.
func (s Stats) Level() int {
	// Binary search to find the level based on XP
	low, high := 0, int(s.Xp)

	for low < high {
		mid := (low + high + 1) / 2
		if s.totalXpForLevel(mid) <= s.Xp {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low
}

func (s Stats) totalXpForLevel(level int) uint64 {
	if level == 0 {
		return 0
	}
	total := uint64(0)

	// Calculate cumulative XP required to reach the specified level
	for i := 0; i < level; i++ {
		total += uint64(math.Pow(float64(i), float64(x))) + C
	}
	return total
}

func (s Stats) UntilNextLevel() uint {
	level := s.Level()
	nextXp := uint64(math.Pow(float64(level), float64(x))) + C
	currentTotalXp := s.totalXpForLevel(level)
	return uint(nextXp - (s.Xp - currentTotalXp))
}
