package repositories

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"live-polling-tool/backend/models"
	"time"
)

type Repo struct{ DB *mongo.Database }

func New(db *mongo.Database) *Repo { return &Repo{DB: db} }
func (r *Repo) Init(ctx context.Context) error {
	_, e := r.DB.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)})
	if e != nil {
		return e
	}
	_, e = r.DB.Collection("votes").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "poll_id", Value: 1}, {Key: "voter_id", Value: 1}}, Options: options.Index().SetUnique(true)})
	return e
}
func (r *Repo) CreateUser(ctx context.Context, u models.User) error {
	_, e := r.DB.Collection("users").InsertOne(ctx, u)
	return e
}
func (r *Repo) FindUser(ctx context.Context, email string) (models.User, error) {
	var u models.User
	e := r.DB.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&u)
	return u, e
}
func (r *Repo) CreatePoll(ctx context.Context, p models.Poll) error {
	_, e := r.DB.Collection("polls").InsertOne(ctx, p)
	return e
}
func (r *Repo) FindPoll(ctx context.Context, id string) (models.Poll, error) {
	var p models.Poll
	e := r.DB.Collection("polls").FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	return p, e
}
func (r *Repo) ListPolls(ctx context.Context, owner string) ([]models.Poll, error) {
	cur, e := r.DB.Collection("polls").Find(ctx, bson.M{"owner_id": owner})
	if e != nil {
		return nil, e
	}
	defer cur.Close(ctx)
	var out []models.Poll
	e = cur.All(ctx, &out)
	return out, e
}
func (r *Repo) UpdatePoll(ctx context.Context, id, owner string, set bson.M) error {
	set["updated_at"] = time.Now()
	res, e := r.DB.Collection("polls").UpdateOne(ctx, bson.M{"_id": id, "owner_id": owner}, bson.M{"$set": set})
	if e == nil && res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return e
}
func (r *Repo) DeletePoll(ctx context.Context, id, owner string) error {
	res, e := r.DB.Collection("polls").DeleteOne(ctx, bson.M{"_id": id, "owner_id": owner})
	if e == nil && res.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return e
}
func (r *Repo) AddVote(ctx context.Context, v models.Vote) (bool, error) {
	_, e := r.DB.Collection("votes").InsertOne(ctx, v)
	if mongo.IsDuplicateKeyError(e) {
		return false, nil
	}
	if e != nil {
		return false, e
	}
	res, e := r.DB.Collection("polls").UpdateOne(ctx, bson.M{"_id": v.PollID, "options.id": v.OptionID}, bson.M{"$inc": bson.M{"options.$.votes": 1}, "$set": bson.M{"updated_at": time.Now()}})
	if e != nil || res.MatchedCount != 1 {
		_, _ = r.DB.Collection("votes").DeleteOne(ctx, bson.M{"_id": v.ID})
		if e != nil {
			return false, e
		}
		return false, mongo.ErrNoDocuments
	}
	return true, nil
}
func (r *Repo) Results(ctx context.Context, id string) (models.Result, error) {
	p, e := r.FindPoll(ctx, id)
	if e != nil {
		return models.Result{}, e
	}
	var total int64
	for _, o := range p.Options {
		total += o.Votes
	}
	return models.Result{PollID: p.ID, Question: p.Question, Options: p.Options, TotalVotes: total}, nil
}
func IsNotFound(e error) bool { return errors.Is(e, mongo.ErrNoDocuments) }
