package room

// PlayerID identifies a connected player. It is assigned by the caller
// (use case layer), not by the room.
type PlayerID string

// Hitbox size in pixels. The box sits at the raccoon's feet: horizontally
// centered on Position, with its bottom edge on Position.Y. A short, wide
// box suits the 3/4 view (RN6): heads may overlap scenery, feet may not.
const (
	HitboxWidth  = 20.0
	HitboxHeight = 12.0
)

// player is the room's internal, mutable view of someone in it.
type player struct {
	id       PlayerID
	nickname Nickname
	position Vec // feet, in room pixels
	input    Input
}

func (p *player) hitbox() Rect {
	return Rect{
		X: p.position.X - HitboxWidth/2,
		Y: p.position.Y - HitboxHeight,
		W: HitboxWidth,
		H: HitboxHeight,
	}
}

func (p *player) setHitbox(b Rect) {
	p.position = Vec{X: b.X + HitboxWidth/2, Y: b.Y + HitboxHeight}
}

func (p *player) view() PlayerView {
	return PlayerView{ID: p.id, Nickname: p.nickname.String(), Position: p.position}
}

// PlayerView is a read-only snapshot of a player, safe to hand to other layers.
type PlayerView struct {
	ID       PlayerID
	Nickname string
	Position Vec
}
