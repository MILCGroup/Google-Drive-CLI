package files

import (
	"context"
	"sort"
	"strings"

	"github.com/milcgroup/gdrv/internal/types"
	"github.com/sahilm/fuzzy"
)

// FuzzySearchOptions configures fuzzy search parameters
type FuzzySearchOptions struct {
	Query      string
	Threshold  int    // Minimum match score (0-100)
	Limit      int    // Maximum number of results
	FolderID   string // Search within specific folder
	MimeType   string // Filter by MIME type
	PageSize   int    // Page size for fetching files
	MaxResults int    // Maximum files to fetch from API
}

// FuzzyMatch represents a file with fuzzy match score
type FuzzyMatch struct {
	File  *types.DriveFile
	Score int
	Match *fuzzy.Match
}

// FuzzySearchResult contains ranked fuzzy search results
type FuzzySearchResult struct {
	Matches    []FuzzyMatch
	TotalFiles int // Total files scanned
	Query      string
}

// FuzzySearch performs fuzzy search across files
func (m *Manager) FuzzySearch(ctx context.Context, reqCtx *types.RequestContext, opts FuzzySearchOptions) (*FuzzySearchResult, error) {
	// Build list options with filters
	listOpts := ListOptions{
		PageSize:       opts.PageSize,
		IncludeTrashed: false,
		Fields:         "id,name,mimeType,size,createdTime,modifiedTime,parents,trashed",
	}

	if opts.FolderID != "" {
		listOpts.ParentID = opts.FolderID
		reqCtx.InvolvedParentIDs = append(reqCtx.InvolvedParentIDs, opts.FolderID)
	}

	// Add MIME type filter to query if specified
	if opts.MimeType != "" {
		listOpts.Query = "mimeType = '" + opts.MimeType + "'"
	}

	// Determine max files to fetch
	maxFiles := opts.MaxResults
	if maxFiles == 0 {
		maxFiles = 1000 // Default limit to avoid fetching too many files
	}

	// Fetch files with pagination
	var allFiles []*types.DriveFile
	totalFiles := 0
	pageToken := ""

	for {
		listOpts.PageToken = pageToken
		result, err := m.List(ctx, reqCtx, listOpts)
		if err != nil {
			return nil, err
		}

		totalFiles += len(result.Files)
		allFiles = append(allFiles, result.Files...)

		// Stop if we've reached the max results or there are no more pages
		if len(allFiles) >= maxFiles || result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	// Apply fuzzy matching
	matches := m.fuzzyFilter(allFiles, opts.Query, opts.Threshold)

	// Sort by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Apply limit
	if opts.Limit > 0 && len(matches) > opts.Limit {
		matches = matches[:opts.Limit]
	}

	return &FuzzySearchResult{
		Matches:    matches,
		TotalFiles: totalFiles,
		Query:      opts.Query,
	}, nil
}

// fuzzyFilter applies fuzzy matching to a slice of files
func (m *Manager) fuzzyFilter(files []*types.DriveFile, pattern string, threshold int) []FuzzyMatch {
	if threshold == 0 {
		threshold = 30 // Default threshold if not specified
	}

	// Convert files to strings for fuzzy matching
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}

	// Find fuzzy matches
	fuzzyMatches := fuzzy.Find(pattern, names)

	// Filter by threshold and build results
	var results []FuzzyMatch
	for _, match := range fuzzyMatches {
		// Calculate score as percentage (0-100)
		// The library returns a score where higher is better, but we want a percentage
		score := calculateMatchScore(match, pattern)

		if score >= threshold {
			results = append(results, FuzzyMatch{
				File:  files[match.Index],
				Score: score,
				Match: &match,
			})
		}
	}

	return results
}

// calculateMatchScore converts the fuzzy match score to a 0-100 percentage
func calculateMatchScore(match fuzzy.Match, pattern string) int {
	// sahilm/fuzzy returns a score where higher is better
	// We normalize it to a 0-100 scale
	// Score is roughly: match length + bonus for consecutive matches + bonus for exact matches

	// Simple percentage based on the match score relative to pattern length
	// This is a heuristic - the actual score depends on the library's scoring algorithm
	if match.Score <= 0 {
		return 0
	}

	// Calculate a percentage based on the score
	// Higher scores indicate better matches
	maxPossibleScore := len(pattern) * 3 // Rough estimate of max score
	percentage := (match.Score * 100) / maxPossibleScore

	if percentage > 100 {
		percentage = 100
	}
	if percentage < 0 {
		percentage = 0
	}

	return percentage
}

// FuzzySearchAll performs fuzzy search across all files in Drive (no folder restriction)
func (m *Manager) FuzzySearchAll(ctx context.Context, reqCtx *types.RequestContext, opts FuzzySearchOptions) (*FuzzySearchResult, error) {
	// Set a reasonable max limit for searching all files
	if opts.MaxResults == 0 {
		opts.MaxResults = 2000
	}

	return m.FuzzySearch(ctx, reqCtx, opts)
}

// FuzzyMatchSource implements fuzzy.Source for custom matching
type FuzzyMatchSource struct {
	files   []*types.DriveFile
	targets []string
}

// String returns the string at the given index
func (s FuzzyMatchSource) String(i int) string {
	return s.targets[i]
}

// Len returns the number of items
func (s FuzzyMatchSource) Len() int {
	return len(s.targets)
}

// FuzzySearchWithSource performs fuzzy search using custom data source
func (m *Manager) FuzzySearchWithSource(ctx context.Context, reqCtx *types.RequestContext, pattern string, source FuzzyMatchSource, threshold int) []FuzzyMatch {
	matches := fuzzy.FindFrom(pattern, source)

	var results []FuzzyMatch
	for _, match := range matches {
		score := calculateMatchScore(match, pattern)
		if score >= threshold {
			results = append(results, FuzzyMatch{
				File:  source.files[match.Index],
				Score: score,
				Match: &match,
			})
		}
	}

	return results
}

// QuickFuzzy performs a quick fuzzy search with minimal options
func (m *Manager) QuickFuzzy(ctx context.Context, reqCtx *types.RequestContext, query string, limit int) ([]FuzzyMatch, error) {
	opts := FuzzySearchOptions{
		Query:      query,
		Limit:      limit,
		PageSize:   100,
		MaxResults: 500,
		Threshold:  30,
	}

	result, err := m.FuzzySearchAll(ctx, reqCtx, opts)
	if err != nil {
		return nil, err
	}

	return result.Matches, nil
}

// BuildMatchPreview creates a preview string with matched characters highlighted
func BuildMatchPreview(filename string, match *fuzzy.Match, maxLength int) string {
	if match == nil {
		return filename
	}

	// If the filename is short enough, show it all
	if len(filename) <= maxLength {
		return highlightMatches(filename, match.MatchedIndexes)
	}

	// Otherwise, try to show the matched section
	if len(match.MatchedIndexes) > 0 {
		start := match.MatchedIndexes[0]
		end := match.MatchedIndexes[len(match.MatchedIndexes)-1]

		// Add some context around the match
		contextStart := start - 10
		if contextStart < 0 {
			contextStart = 0
		}
		contextEnd := end + 10
		if contextEnd > len(filename) {
			contextEnd = len(filename)
		}

		prefix := ""
		suffix := ""
		if contextStart > 0 {
			prefix = "..."
		}
		if contextEnd < len(filename) {
			suffix = "..."
		}

		// Adjust indexes for the substring
		adjustedIndexes := make([]int, len(match.MatchedIndexes))
		for i, idx := range match.MatchedIndexes {
			adjustedIndexes[i] = idx - contextStart + len(prefix)
		}

		substring := filename[contextStart:contextEnd]
		return prefix + highlightMatches(substring, adjustedIndexes) + suffix
	}

	// Fallback: just show the beginning
	return filename[:maxLength-3] + "..."
}

// highlightMatches wraps matched characters in brackets
func highlightMatches(str string, indexes []int) string {
	if len(indexes) == 0 {
		return str
	}

	var result strings.Builder
	idxMap := make(map[int]bool)
	for _, idx := range indexes {
		idxMap[idx] = true
	}

	for i, ch := range str {
		if idxMap[i] {
			result.WriteString("[")
			result.WriteRune(ch)
			result.WriteString("]")
		} else {
			result.WriteRune(ch)
		}
	}

	return result.String()
}
