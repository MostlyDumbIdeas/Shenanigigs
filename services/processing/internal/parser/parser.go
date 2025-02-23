package parser

import (
	"encoding/json"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"
	"time"

	"shenanigigs/processing/internal/models"
)

type RawJobPosting struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PostedAt    time.Time `json:"posted_at"`
	RawText     string    `json:"raw_text"`
	ParentID    int       `json:"parent_id"`
}

var (
	companyPattern    = regexp.MustCompile(`(?i)(company|at):\s*([^,|\n]+)`)
	locationPattern   = regexp.MustCompile(`(?i)(location|remote):\s*([^,|\n]+)`)
	salaryPattern     = regexp.MustCompile(`\$(\d+(?:,\d{3})*(?:\.\d{2})?)[Kk]?\s*-\s*\$(\d+(?:,\d{3})*(?:\.\d{2})?)[Kk]?`)
	titlePattern      = regexp.MustCompile(`(?i)(position|role|title):\s*([^,|\n]+)`)
	remotePattern     = regexp.MustCompile(`(?i)(remote|wfh|work[- ]from[- ]home)`)
	techStackPattern  = regexp.MustCompile(`(?i)(tech stack|technologies|skills):\s*([^,|\n]+)`)
	experiencePattern = regexp.MustCompile(`(?i)(experience|yoe|years):\s*([^,|\n]+)`)
)

func generateUUIDFromID(id string) string {
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	uuid := uuid.NewSHA1(namespace, []byte(id))
	return uuid.String()
}

func ParseJobPosting(rawData string) (*models.JobPosting, error) {
	var raw RawJobPosting
	if err := json.Unmarshal([]byte(rawData), &raw); err != nil {
		return nil, err
	}

	uuidStr := generateUUIDFromID(raw.ID)

	cleanText := normalizeText(raw.RawText)
	titleParts := strings.Split(cleanText, " | ")

	company := titleParts[0]
	location := ""
	title := ""
	experienceLevel := ""
	technologies := []string{}

	for i, part := range titleParts {
		switch i {
		case 1:
			location = strings.TrimSpace(part)
		case 2:
			titleParts := strings.Split(part, " ")
			experienceLevel = titleParts[0]
			title = strings.Join(titleParts[1:], " ")
		case 4:
			techStr := strings.TrimSpace(part)
			technologies = strings.Split(techStr, ", ")
		}
	}

	if title == "" {
		if matches := titlePattern.FindStringSubmatch(raw.Description); len(matches) > 2 {
			title = strings.TrimSpace(matches[2])
		}
	}

	if company == "" {
		if matches := companyPattern.FindStringSubmatch(raw.Description); len(matches) > 2 {
			company = strings.TrimSpace(matches[2])
		}
	}

	if location == "" {
		if matches := locationPattern.FindStringSubmatch(raw.Description); len(matches) > 2 {
			location = strings.TrimSpace(matches[2])
		}
	}

	if len(technologies) == 0 {
		technologies = extractTechnologies(raw.Description)
	}

	if experienceLevel == "" {
		experienceLevel = extractExperienceLevel(raw.Description)
	}

	compMatches := salaryPattern.FindStringSubmatch(raw.Description)
	var compMin, compMax float64

	if len(compMatches) > 2 {
		minStr := strings.ReplaceAll(compMatches[1], ",", "")
		maxStr := strings.ReplaceAll(compMatches[2], ",", "")

		if strings.HasSuffix(minStr, "k") || strings.HasSuffix(minStr, "K") {
			minStr = strings.TrimSuffix(strings.TrimSuffix(minStr, "k"), "K")
			compMin = parseFloat(minStr) * 1000
		} else {
			compMin = parseFloat(minStr)
		}

		if strings.HasSuffix(maxStr, "k") || strings.HasSuffix(maxStr, "K") {
			maxStr = strings.TrimSuffix(strings.TrimSuffix(maxStr, "k"), "K")
			compMax = parseFloat(maxStr) * 1000
		} else {
			compMax = parseFloat(maxStr)
		}
	}

	remotePolicy := "unknown"
	if strings.Contains(strings.ToLower(location), "remote") || remotePattern.MatchString(raw.Description) {
		remotePolicy = "remote"
	}

	return &models.JobPosting{
		ID:                   uuidStr,
		Title:                title,
		Company:              company,
		Location:             location,
		Description:          raw.Description,
		Technologies:         technologies,
		ExperienceLevel:      experienceLevel,
		CompensationMin:      compMin,
		CompensationMax:      compMax,
		CompensationCurrency: "USD",
		CompensationPeriod:   "yearly",
		RemotePolicy:         remotePolicy,
		Source:               "hackernews",
		SourceURL:            "",
		CreatedAt:            raw.PostedAt,
		UpdatedAt:            time.Now(),
		RawData:              rawData,
	}, nil
}

func normalizeText(text string) string {
	text = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`[\x{2013}\x{2014}\x{2015}]`).ReplaceAllString(text, "-")
	return strings.TrimSpace(text)
}

func extractTechnologies(text string) []string {
	var technologies []string

	if matches := techStackPattern.FindStringSubmatch(text); len(matches) > 2 {
		techText := matches[2]
		technologies = strings.Split(techText, ",")
	}

	commonTech := []string{
		"python", "javascript", "typescript", "java", "golang", "ruby", "php",
		"react", "angular", "vue", "node", "django", "flask", "spring",
		"aws", "azure", "gcp", "kubernetes", "docker", "terraform",
		"sql", "mongodb", "postgresql", "mysql", "redis",
	}

	textLower := strings.ToLower(text)
	for _, tech := range commonTech {
		if strings.Contains(textLower, tech) {
			technologies = append(technologies, tech)
		}
	}

	seen := make(map[string]bool)
	var result []string

	for _, tech := range technologies {
		tech = strings.ToLower(strings.TrimSpace(tech))
		if tech != "" && !seen[tech] {
			seen[tech] = true
			result = append(result, tech)
		}
	}

	return result
}

func extractExperienceLevel(text string) string {
	if matches := experiencePattern.FindStringSubmatch(text); len(matches) > 2 {
		exp := strings.ToLower(matches[2])
		if strings.Contains(exp, "senior") || strings.Contains(exp, "sr") || strings.Contains(exp, "5+") {
			return "Senior"
		} else if strings.Contains(exp, "junior") || strings.Contains(exp, "jr") || strings.Contains(exp, "entry") {
			return "Junior"
		} else if strings.Contains(exp, "mid") || strings.Contains(exp, "intermediate") {
			return "Mid-Level"
		}
	}

	text = strings.ToLower(text)
	if strings.Contains(text, "senior") || strings.Contains(text, "sr.") || strings.Contains(text, "lead") {
		return "Senior"
	} else if strings.Contains(text, "junior") || strings.Contains(text, "jr.") || strings.Contains(text, "entry") {
		return "Junior"
	} else if strings.Contains(text, "mid") || strings.Contains(text, "intermediate") {
		return "Mid-Level"
	}

	return "Not Specified"
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
