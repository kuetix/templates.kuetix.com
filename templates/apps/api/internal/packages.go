package internal

import (
	"errors"
	"fmt"

	"github.com/kuetix/uuid"
)

//goland:noinspection GoUnusedConst
const (
	PackageVisibilityTypePublic  = "public"
	PackageVisibilityTypePrivate = "private"
	PackageTypeModule            = "module"
	PackageTypePackage           = "package"
	PackageTypeApplication       = "application"
	PackageTypeLibrary           = "library"
	PackageTypeFramework         = "framework"
	PackageTypeTool              = "tool"
	PackageTypeService           = "service"
	PackageTypeDatabase          = "database"
	PackageTypeMiddleware        = "middleware"
	PackageTypeUtility           = "utility"
)

type OwnerPackages struct {
	Owner   string                      `json:"owner"`
	Package map[string]PackageOwnerMeta `json:"packages"`
}

type PackageOwnerMeta struct {
	ID         string                `json:"id"`
	Owner      string                `json:"owner"`
	Type       string                `json:"type"`
	Name       string                `json:"name"`
	Path       string                `json:"path"`
	Hash       string                `json:"hash"`
	Visibility PackageVisibility     `json:"visibility"`
	Versions   []PackageOwnerVersion `json:"versions"`
}

type PackageOwnerVersion struct {
	Version    string            `json:"version"`
	Hash       string            `json:"hash"`
	Visibility PackageVisibility `json:"visibility"`
}

type PackageBasic struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Owner       string `json:"owner"`
	Description string `json:"description,omitempty"`
	License     string `json:"license"`
}

type PackageAuthors struct {
	Author       string   `json:"author,omitempty"`
	Maintainers  []string `json:"maintainers,omitempty"`
	Contributors []string `json:"contributors,omitempty"`
}

type PackageURL struct {
	License       string `json:"license,omitempty"`
	Homepage      string `json:"homepage,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	Repository    string `json:"repository,omitempty"`
	Terms         string `json:"terms,omitempty"`
	Privacy       string `json:"privacy,omitempty"`
	Support       string `json:"support,omitempty"`
	Bugs          string `json:"bugs,omitempty"`
	Changelog     string `json:"changelog,omitempty"`
	Issues        string `json:"issues,omitempty"`
	PullRequests  string `json:"pullRequests,omitempty"`
	Commits       string `json:"commits,omitempty"`
}

type PackageCategories struct {
	Keywords   []string `json:"keywords,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type PackageDependencies struct {
	Packages        string `json:"packages,omitempty"`
	Count           int    `json:"count,omitempty"`
	DependentsCount int    `json:"dependentsCount,omitempty"`
}

type PackageAnalytics struct {
	Downloads int `json:"downloads,omitempty"`
	Likes     int `json:"likes,omitempty"`
	Followers int `json:"followers,omitempty"`
	Stars     int `json:"stars,omitempty"`
}

type PackageStatus struct {
	BuildStatus string `json:"buildStatus,omitempty"`
	Status      string `json:"status,omitempty"`
}

type PackageVisibility struct {
	IsPublished bool `json:"isPublished,omitempty"`
	IsPrivate   bool `json:"isPrivate,omitempty"`
}

type PackageQuality struct {
	QualityScore  int     `json:"qualityScore,omitempty"`
	SecurityScore int     `json:"securityScore,omitempty"`
	TestCoverage  float64 `json:"testCoverage,omitempty"`
}

type PackageDate struct {
	CreatedAt string `json:"createdAt,omitempty"`
	PublishAt string `json:"publishAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// PricingTier represents pricing information for a package
type PricingTier struct {
	OneTimePrice      float64 `json:"oneTimePrice"`      // One-time purchase price
	SubscriptionPrice float64 `json:"subscriptionPrice"` // Monthly subscription price
	CloudPrice        float64 `json:"cloudPrice"`        // Cloud-hosted version price
	EnterprisePrice   float64 `json:"enterprisePrice"`   // Custom enterprise pricing
	Currency          string  `json:"currency"`          // Currency code (e.g., USD, EUR)
}

// Package represents a package in the registry
type Package struct {
	ID           string              `json:"id,omitempty"`
	Basic        PackageBasic        `json:"basic"`
	Authors      PackageAuthors      `json:"authors,omitempty"`
	URL          PackageURL          `json:"url"`
	Categories   PackageCategories   `json:"categories,omitempty"`
	Dependencies PackageDependencies `json:"dependencies,omitempty"`
	Analytics    PackageAnalytics    `json:"analytics,omitempty"`
	Status       PackageStatus       `json:"status,omitempty"`
	Visibility   PackageVisibility   `json:"visibility,omitempty"`
	Quality      PackageQuality      `json:"quality,omitempty"`
	Pricing      PricingTier         `json:"pricing,omitempty"`
	Date         PackageDate         `json:"date,omitempty"`
	Modules      []string            `json:"modules,omitempty"`
}

func (op *OwnerPackages) GetVersion(version string) (*PackageOwnerMeta, error) {
	if version == "" {
		return nil, errors.New("version cannot be empty")
	}

	meta, ok := op.Package[version]

	if !ok {
		return nil, fmt.Errorf("version %s not found for owner %s", version, op.Owner)
	}

	return &meta, nil
}

func (op *OwnerPackages) GetVersionByHash(hash string) (*PackageOwnerMeta, error) {
	if hash == "" {
		return nil, errors.New("hash cannot be empty")
	}

	for _, version := range op.Package {
		if version.Hash == hash {
			return &version, nil
		}
	}

	return nil, fmt.Errorf("version with hash %s not found for owner %s", hash, op.Owner)
}

func (op *OwnerPackages) GetVersionFromString(source string) string {
	id := uuid.Base64Id(source)

	return id
}
