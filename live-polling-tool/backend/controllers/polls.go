package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"live-polling-tool/backend/models"
	"live-polling-tool/backend/realtime"
	"live-polling-tool/backend/repositories"
	"live-polling-tool/backend/utils"
	"net/http"
	"strings"
	"time"
)

type PollController struct {
	Repo   *repositories.Repo
	Redis  *realtime.Redis
	Secure bool
}
type pollReq struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

func (p *PollController) Create(c *gin.Context) {
	var r pollReq
	if c.ShouldBindJSON(&r) != nil {
		utils.Error(c, 400, "INVALID_REQUEST", "Valid JSON is required")
		return
	}
	q := strings.TrimSpace(r.Question)
	if len(q) < 5 || len(q) > 240 || len(r.Options) < 2 || len(r.Options) > 10 {
		utils.Error(c, 400, "VALIDATION_ERROR", "Question must be 5-240 characters and poll must have 2-10 options")
		return
	}
	seen := map[string]bool{}
	opts := make([]models.PollOption, 0, len(r.Options))
	for _, x := range r.Options {
		x = strings.TrimSpace(x)
		key := strings.ToLower(x)
		if len(x) < 1 || len(x) > 120 || seen[key] {
			utils.Error(c, 400, "VALIDATION_ERROR", "Options must be unique and 1-120 characters")
			return
		}
		seen[key] = true
		opts = append(opts, models.PollOption{ID: uuid.NewString(), Text: x})
	}
	now := time.Now()
	poll := models.Poll{ID: uuid.NewString(), Question: q, Options: opts, OwnerID: c.GetString("userID"), Status: "active", CreatedAt: now, UpdatedAt: now}
	if e := p.Repo.CreatePoll(c.Request.Context(), poll); e != nil {
		utils.Error(c, 500, "DB_ERROR", "Could not create poll")
		return
	}
	utils.OK(c, poll)
}
func (p *PollController) List(c *gin.Context) {
	ps, e := p.Repo.ListPolls(c.Request.Context(), c.GetString("userID"))
	if e != nil {
		utils.Error(c, 500, "DB_ERROR", "Could not load polls")
		return
	}
	utils.OK(c, ps)
}
func (p *PollController) Get(c *gin.Context) {
	poll, e := p.Repo.FindPoll(c.Request.Context(), c.Param("id"))
	if e != nil || poll.Status == "" {
		utils.Error(c, 404, "NOT_FOUND", "Poll not found")
		return
	}
	utils.OK(c, publicPoll(poll))
}
func (p *PollController) Results(c *gin.Context) {
	r, e := p.Repo.Results(c.Request.Context(), c.Param("id"))
	if e != nil {
		utils.Error(c, 404, "NOT_FOUND", "Poll not found")
		return
	}
	utils.OK(c, r)
}
func (p *PollController) Vote(c *gin.Context) {
	var body struct {
		OptionID string `json:"optionId"`
	}
	if c.ShouldBindJSON(&body) != nil || body.OptionID == "" {
		utils.Error(c, 400, "INVALID_REQUEST", "optionId is required")
		return
	}
	poll, e := p.Repo.FindPoll(c.Request.Context(), c.Param("id"))
	if e != nil {
		utils.Error(c, 404, "NOT_FOUND", "Poll not found")
		return
	}
	if poll.Status != "active" {
		utils.Error(c, 409, "POLL_CLOSED", "This poll is closed")
		return
	}
	valid := false
	for _, o := range poll.Options {
		if o.ID == body.OptionID {
			valid = true
		}
	}
	if !valid {
		utils.Error(c, 400, "INVALID_OPTION", "Option does not belong to this poll")
		return
	}
	voter, err := c.Cookie("voter_id")
	if err != nil || len(voter) < 20 {
		voter = uuid.NewString()
		if p.Secure {
			c.SetSameSite(http.SameSiteNoneMode)
		} else {
			c.SetSameSite(http.SameSiteLaxMode)
		}
		c.SetCookie("voter_id", voter, 31536000, "/", "", p.Secure, true)
	}
	added, e := p.Repo.AddVote(c.Request.Context(), models.Vote{ID: uuid.NewString(), PollID: poll.ID, OptionID: body.OptionID, VoterID: voter, CreatedAt: time.Now()})
	if e != nil {
		utils.Error(c, 500, "VOTE_FAILED", "Could not submit vote")
		return
	}
	if !added {
		utils.Error(c, 409, "ALREADY_VOTED", "You have already voted on this poll")
		return
	}
	res, e := p.Repo.Results(c.Request.Context(), poll.ID)
	if e == nil && p.Redis != nil {
		_ = p.Redis.PublishResult(c.Request.Context(), res)
	}
	utils.OK(c, res)
}
func (p *PollController) Close(c *gin.Context) {
	if e := p.Repo.UpdatePoll(c.Request.Context(), c.Param("id"), c.GetString("userID"), map[string]any{"status": "closed"}); e != nil {
		utils.Error(c, 404, "NOT_FOUND", "Poll not found")
		return
	}
	utils.OK(c, gin.H{"status": "closed"})
}
func (p *PollController) Delete(c *gin.Context) {
	if e := p.Repo.DeletePoll(c.Request.Context(), c.Param("id"), c.GetString("userID")); e != nil {
		utils.Error(c, 404, "NOT_FOUND", "Poll not found")
		return
	}
	c.Status(http.StatusNoContent)
}
func publicPoll(p models.Poll) gin.H {
	return gin.H{"id": p.ID, "question": p.Question, "options": p.Options, "status": p.Status, "createdAt": p.CreatedAt}
}
