package transitions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	jsonDbCache "github.com/anare/simple-json-db-cache"
	"github.com/kuetix/engine/engine/domain"
	"github.com/kuetix/engine/engine/domain/interfaces"
	"github.com/kuetix/engine/engine/workflow"
)

type profilesTransitions struct {
	workflow.BaseServiceTransition
	dbm   *jsonDbCache.DB
	items *jsonDbCache.Collection
}

func NewProfilesTransitions() interfaces.ServiceTransitions {
	return &profilesTransitions{}
}

// Profile represents a user profile in the system
type Profile struct {
	ID          string            `json:"id"`
	UserID      string            `json:"userId"` // User ID from JWT token
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	FullName    string            `json:"fullName"`
	Bio         string            `json:"bio"`
	Avatar      string            `json:"avatar"`
	Location    string            `json:"location"`
	Website     string            `json:"website"`
	Company     string            `json:"company"`
	SocialLinks map[string]string `json:"socialLinks"` // e.g., {"github": "username", "twitter": "handle"}
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

// getDB initializes and returns the database connection
func (p *profilesTransitions) getDB() (*jsonDbCache.Collection, error) {
	if p.items != nil {
		return p.items, nil
	}

	options := p.Ctx.Engine.GetApplication().Env.Options
	dbPath := options.Context["dbPath"].(string)

	// Default path if not provided
	if dbPath == "" {
		dbPath = "./runtime/data"
	}

	// Ensure directory exists
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dbFile := filepath.Join(dbPath, "profile")
	db, err := jsonDbCache.NewDB(dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	p.dbm = db
	p.items = p.dbm.NewCollection("profile")
	return p.items, nil
}

// CreateProfile creates a new profile
func (p *profilesTransitions) CreateProfile(args map[string]interface{}) (r domain.FlowStepResult) {
	// Get database connection
	db, err := p.getDB()
	if err != nil {
		r.Success = false
		r.Error = err
		return
	}

	// Extract profile data from args
	userID, _ := args["userId"].(string)
	if userID == "" {
		r.Success = false
		r.Error = fmt.Errorf("userId is required")
		return
	}

	// Check if profile already exists for this user
	profileID := fmt.Sprintf("profile_%s", userID)
	if db.Exists(profileID) {
		r.Success = false
		r.Error = fmt.Errorf("profile already exists for user %s", userID)
		return
	}

	// Create new profile
	now := time.Now().Format(time.RFC3339)
	profile := Profile{
		ID:        profileID,
		UserID:    userID,
		Username:  getStringArg(args, "username"),
		Email:     getStringArg(args, "email"),
		FullName:  getStringArg(args, "fullName"),
		Bio:       getStringArg(args, "bio"),
		Avatar:    getStringArg(args, "avatar"),
		Location:  getStringArg(args, "location"),
		Website:   getStringArg(args, "website"),
		Company:   getStringArg(args, "company"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Handle social links
	if socialLinks, ok := args["socialLinks"].(map[string]interface{}); ok {
		profile.SocialLinks = make(map[string]string)
		for key, val := range socialLinks {
			if strVal, ok := val.(string); ok {
				profile.SocialLinks[key] = strVal
			}
		}
	}

	// Save to database
	if err := db.Set(profileID, profile); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to save profile: %w", err)
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"profile": profile,
		"message": "Profile created successfully",
	}

	return
}

// GetProfile retrieves a profile by user ID
func (p *profilesTransitions) GetProfile(userId string) (r domain.FlowStepResult) {
	// Get database connection
	db, err := p.getDB()
	if err != nil {
		r.Success = false
		r.Error = err
		return
	}

	// Get user ID
	if userId == "" {
		r.Success = false
		r.Error = fmt.Errorf("userId is required")
		return
	}

	// Retrieve profile
	// profileID: = fmt.Sprintf("profile_%s", userId)
	profileID := userId
	var profile Profile
	if err := db.Get(profileID, &profile); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("profile not found")
		return
	}

	r.Success = true
	r.Response = profile
	return
}

// UpdateProfile updates an existing profile
func (p *profilesTransitions) UpdateProfile(userId string, args map[string]interface{}) (r domain.FlowStepResult) {
	// Get database connection
	db, err := p.getDB()
	if err != nil {
		r.Success = false
		r.Error = err
		return
	}

	// Get user ID
	// userID, _ := args["userId"].(string)
	if userId == "" {
		r.Success = false
		r.Error = fmt.Errorf("userId is required")
		return
	}

	// Retrieve existing profile
	// profileID := fmt.Sprintf("profile_%s", userId)
	profileID := userId
	var profile Profile
	if err := db.Get(profileID, &profile); err != nil {
		profile = Profile{ID: profileID, UserID: userId}
	}

	// Update fields if provided
	if username, ok := args["username"].(string); ok && username != "" {
		profile.Username = username
	}
	if email, ok := args["email"].(string); ok && email != "" {
		profile.Email = email
	}
	if fullName, ok := args["fullName"].(string); ok && fullName != "" {
		profile.FullName = fullName
	}
	if bio, ok := args["bio"].(string); ok {
		profile.Bio = bio
	}
	if avatar, ok := args["avatar"].(string); ok {
		profile.Avatar = avatar
	}
	if location, ok := args["location"].(string); ok {
		profile.Location = location
	}
	if website, ok := args["website"].(string); ok {
		profile.Website = website
	}
	if company, ok := args["company"].(string); ok {
		profile.Company = company
	}

	// Handle social links update
	if socialLinks, ok := args["socialLinks"].(map[string]interface{}); ok {
		if profile.SocialLinks == nil {
			profile.SocialLinks = make(map[string]string)
		}
		for key, val := range socialLinks {
			if strVal, ok := val.(string); ok {
				profile.SocialLinks[key] = strVal
			}
		}
	}

	// Update timestamp
	profile.UpdatedAt = time.Now().Format(time.RFC3339)

	// Update profile in database
	if err := db.Update(profileID, profile); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to update profile: %w", err)
		return
	}

	// "message": "Profile updated successfully",
	r.Success = true
	r.Response = profile
	return
}

// DeleteProfile deletes a profile
func (p *profilesTransitions) DeleteProfile(args map[string]interface{}) (r domain.FlowStepResult) {
	// Get database connection
	db, err := p.getDB()
	if err != nil {
		r.Success = false
		r.Error = err
		return
	}

	// Get user ID
	userID, _ := args["userId"].(string)
	if userID == "" {
		r.Success = false
		r.Error = fmt.Errorf("userId is required")
		return
	}

	// Check if profile exists
	profileID := fmt.Sprintf("profile_%s", userID)
	if !db.Exists(profileID) {
		r.Success = false
		r.Error = fmt.Errorf("profile not found")
		return
	}

	// Delete profile
	if err := db.Delete(profileID); err != nil {
		r.Success = false
		r.Error = fmt.Errorf("failed to delete profile: %w", err)
		return
	}

	r.Success = true
	r.Response = map[string]interface{}{
		"message": "Profile deleted successfully",
	}

	return
}

// ListProfiles lists all profiles (optionally with pagination)
func (p *profilesTransitions) ListProfiles() (r domain.FlowStepResult) {
	// Get database connection
	db, err := p.getDB()
	if err != nil {
		r.Success = false
		r.Error = err
		return
	}

	// Get all profiles from database
	allData := db.GetAll()
	profiles := make([]Profile, 0)

	for _, data := range allData {
		var profile Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}

	// Sort by creation date (newest first)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].CreatedAt > profiles[j].CreatedAt
	})

	r.Success = true
	r.Response = map[string]interface{}{
		"profiles": profiles,
		"count":    len(profiles),
	}
	return
}

// Helper function to get string argument
func getStringArg(args map[string]interface{}, key string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return ""
}
