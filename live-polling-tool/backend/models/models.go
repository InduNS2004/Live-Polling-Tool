package models

import "time"

type User struct {
	ID           string    `bson:"_id" json:"id"`
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	CreatedAt    time.Time `bson:"created_at" json:"createdAt"`
}

type PollOption struct {
	ID    string `bson:"id" json:"id"`
	Text  string `bson:"text" json:"text"`
	Votes int64  `bson:"votes" json:"votes"`
}

type Poll struct {
	ID        string       `bson:"_id" json:"id"`
	Question  string       `bson:"question" json:"question"`
	Options   []PollOption `bson:"options" json:"options"`
	OwnerID   string       `bson:"owner_id" json:"ownerId"`
	Status    string       `bson:"status" json:"status"`
	CreatedAt time.Time    `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time    `bson:"updated_at" json:"updatedAt"`
}

type Vote struct {
	ID        string    `bson:"_id"`
	PollID    string    `bson:"poll_id"`
	OptionID  string    `bson:"option_id"`
	VoterID   string    `bson:"voter_id"`
	CreatedAt time.Time `bson:"created_at"`
}

type Result struct {
	PollID     string       `json:"pollId"`
	Question   string       `json:"question"`
	Options    []PollOption `json:"options"`
	TotalVotes int64        `json:"totalVotes"`
}
