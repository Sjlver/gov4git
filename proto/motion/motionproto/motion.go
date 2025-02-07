package motionproto

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gov4git/gov4git/v2/proto/member"
	"github.com/gov4git/gov4git/v2/proto/motion"
	"github.com/gov4git/lib4git/must"
)

type MotionType string

const (
	MotionConcernType  MotionType = "concern"
	MotionProposalType MotionType = "proposal"
)

func ParseMotionType(ctx context.Context, s string) MotionType {
	switch s {
	case string(MotionConcernType):
		return MotionConcernType
	case string(MotionProposalType):
		return MotionProposalType
	}
	must.Panic(ctx, fmt.Errorf("unknown motion type"))
	return MotionType("")
}

type Motion struct {
	OpenedAt time.Time `json:"opened_at"`
	ClosedAt time.Time `json:"closed_at"`
	// instance, immutable
	ID     MotionID          `json:"id"`
	Type   MotionType        `json:"type"`
	Policy motion.PolicyName `json:"policy"`
	Author member.User       `json:"author"` // community user or empty string
	// meta, mutable
	TrackerURL string   `json:"tracker_url"` // link to concern on an external concern tracker, such as a GitHub issue
	Title      string   `json:"title"`
	Body       string   `json:"description"`
	Labels     []string `json:"labels"`
	// state, mutable
	Frozen    bool `json:"frozen"`
	Closed    bool `json:"closed"`
	Cancelled bool `json:"cancelled"`
	//
	Archived bool `json:"archived"`
	// attention ranking, mutable
	Score Score `json:"score"`
	// network, mutable
	RefBy Refs `json:"ref_by"`
	RefTo Refs `json:"ref_to"`
}

func (m Motion) IsConcern() bool {
	return m.Type == MotionConcernType
}

func (m Motion) IsProposal() bool {
	return m.Type == MotionProposalType
}

func (m Motion) GithubArticle() string {
	switch m.Type {
	case MotionConcernType:
		return "an"
	case MotionProposalType:
		return "a"
	default:
		return "an"
	}
}

func (m Motion) GithubType() string {
	switch m.Type {
	case MotionConcernType:
		return "issue"
	case MotionProposalType:
		return "PR"
	default:
		return "issue/PR"
	}
}

func (m Motion) RefersTo(toID MotionID, typ RefType) bool {
	for _, ref := range m.RefTo {
		if ref.To == toID && ref.Type == typ {
			return true
		}
	}
	return false
}

func (m Motion) ReferredBy(fromID MotionID, typ RefType) bool {
	for _, ref := range m.RefBy {
		if ref.From == fromID && ref.Type == typ {
			return true
		}
	}
	return false
}

func (m *Motion) AddRefTo(ref Ref) {
	if !m.RefersTo(ref.To, ref.Type) {
		m.RefTo = append(m.RefTo, ref)
	}
	m.RefTo.Sort()
}

func (m *Motion) AddRefBy(ref Ref) {
	if !m.ReferredBy(ref.From, ref.Type) {
		m.RefBy = append(m.RefBy, ref)
	}
	m.RefBy.Sort()
}

func (m *Motion) RemoveRef(unref Ref) {
	m.RefTo = m.RefTo.Remove(unref)
	m.RefBy = m.RefBy.Remove(unref)
}

type Motions []Motion

func (x Motions) FindID(id MotionID) (Motion, bool) {
	for _, m := range x {
		if m.ID == id {
			return m, true
		}
	}
	return Motion{}, false
}

func (x Motions) Sort() { sort.Sort(x) }

func (x Motions) Len() int { return len(x) }

func (x Motions) Less(i, j int) bool { return x[i].Score.Attention < x[j].Score.Attention }

func (x Motions) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

type MotionsByID []Motion

func (x MotionsByID) Sort() { sort.Sort(x) }

func (x MotionsByID) Len() int { return len(x) }

func (x MotionsByID) Less(i, j int) bool { return x[i].ID < x[j].ID }

func (x MotionsByID) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func SelectOpenMotions(ms Motions) Motions {
	r := Motions{}
	for _, m := range ms {
		if !m.Closed {
			r = append(r, m)
		}
	}
	return r
}
