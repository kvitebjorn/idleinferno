package model

type Stats struct {
	Xp      uint64
	Created string
	Online  bool
}

func (s Stats) Level() int {
	return int(s.Xp / 2) // TODO: an actual levelling formula based on xp
}
