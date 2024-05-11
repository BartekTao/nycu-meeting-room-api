// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Connection interface {
	IsConnection()
	GetEdges() []Edge
	GetPageInfo() *PageInfo
}

type Edge interface {
	IsEdge()
	GetNode() Node
	GetCursor() string
}

type Node interface {
	IsNode()
	GetID() string
}

type Booking struct {
	StartAt  int   `json:"startAt"`
	EndAt    int   `json:"endAt"`
	BookedBy *User `json:"bookedBy"`
}

type Event struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	StartAt      int     `json:"startAt"`
	EndAt        int     `json:"endAt"`
	Room         *Room   `json:"room,omitempty"`
	Participants []*User `json:"participants,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	RemindAt     int     `json:"remindAt"`
	Creator      *User   `json:"creator"`
	IsDelete     *bool   `json:"isDelete,omitempty"`
}

type Mutation struct {
}

type PageInfo struct {
	StartCursor *string `json:"startCursor,omitempty"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

type Query struct {
}

type Room struct {
	ID        string     `json:"id"`
	RoomID    string     `json:"roomId"`
	Capacity  int        `json:"capacity"`
	Equipment []string   `json:"equipment,omitempty"`
	Rules     []string   `json:"rules,omitempty"`
	IsDelete  *bool      `json:"isDelete,omitempty"`
	Bookings  []*Booking `json:"bookings,omitempty"`
}

func (Room) IsNode()            {}
func (this Room) GetID() string { return this.ID }

type RoomConnection struct {
	Edges    []*RoomEdge `json:"edges,omitempty"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

func (RoomConnection) IsConnection() {}
func (this RoomConnection) GetEdges() []Edge {
	if this.Edges == nil {
		return nil
	}
	interfaceSlice := make([]Edge, 0, len(this.Edges))
	for _, concrete := range this.Edges {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}
func (this RoomConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type RoomEdge struct {
	Node   *Room  `json:"node,omitempty"`
	Cursor string `json:"cursor"`
}

func (RoomEdge) IsEdge()                {}
func (this RoomEdge) GetNode() Node     { return *this.Node }
func (this RoomEdge) GetCursor() string { return this.Cursor }

type UpsertEventInput struct {
	ID              *string  `json:"id,omitempty"`
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	StartAt         int      `json:"startAt"`
	EndAt           int      `json:"endAt"`
	RoomID          *string  `json:"roomId,omitempty"`
	ParticipantsIDs []string `json:"participantsIDs,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	RemindAt        int      `json:"remindAt"`
}

type UpsertRoomInput struct {
	ID        *string  `json:"id,omitempty"`
	RoomID    string   `json:"roomId"`
	Capacity  int      `json:"capacity"`
	Equipment []string `json:"equipment,omitempty"`
	Rules     []string `json:"rules,omitempty"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
