package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// SetTargetImport handles CSV target import functionality.
type SetTargetImport struct{}

// NewSetTargetImport creates a new SetTargetImport.
func NewSetTargetImport() *SetTargetImport {
	return &SetTargetImport{}
}

// ParseCSV parses a CSV file and returns target definitions.
// Expected columns: Label, Year, MintMark, Denomination, Country, Material
// Only Label is required; other columns are optional.
func (s *SetTargetImport) ParseCSV(reader io.Reader) ([]models.CoinSetTarget, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read header
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to indices
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Validate required columns
	labelIdx, hasLabel := columnMap["label"]
	if !hasLabel {
		return nil, fmt.Errorf("CSV must contain a 'Label' column")
	}

	yearIdx, _ := columnMap["year"]
	mintIdx, _ := columnMap["mintmark"]
	denomIdx, _ := columnMap["denomination"]
	countryIdx, _ := columnMap["country"]
	materialIdx, _ := columnMap["material"]

	targets := []models.CoinSetTarget{}
	sortOrder := 0
	seenLabels := make(map[string]bool)

	// Read data rows
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(record) <= labelIdx {
			continue
		}

		label := strings.TrimSpace(record[labelIdx])
		if label == "" {
			continue
		}

		// Check for duplicate labels
		labelKey := strings.ToLower(label)
		if seenLabels[labelKey] {
			return nil, fmt.Errorf("duplicate target label: %s", label)
		}
		seenLabels[labelKey] = true

		target := models.CoinSetTarget{
			Label:     label,
			SortOrder: sortOrder,
		}

		// Parse optional year
		if yearIdx < len(record) && record[yearIdx] != "" {
			year, err := strconv.Atoi(strings.TrimSpace(record[yearIdx]))
			if err != nil {
				return nil, fmt.Errorf("invalid year for target '%s': %s", label, record[yearIdx])
			}
			target.Year = &year
		}

		// Parse optional mint mark
		if mintIdx < len(record) && record[mintIdx] != "" {
			mint := strings.TrimSpace(record[mintIdx])
			target.MintMark = &mint
		}

		// Parse optional denomination
		if denomIdx < len(record) && record[denomIdx] != "" {
			denom := strings.TrimSpace(record[denomIdx])
			target.Denomination = &denom
		}

		// Parse optional country
		if countryIdx < len(record) && record[countryIdx] != "" {
			country := strings.TrimSpace(record[countryIdx])
			target.Country = &country
		}

		// Parse optional material
		if materialIdx < len(record) && record[materialIdx] != "" {
			material := strings.TrimSpace(record[materialIdx])
			target.Material = &material
		}

		targets = append(targets, target)
		sortOrder++
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("CSV contains no valid target rows")
	}

	return targets, nil
}

// ValidateTargets checks for duplicate target identities within a set.
func (s *SetTargetImport) ValidateTargets(targets []models.CoinSetTarget) error {
	seen := make(map[string]bool)

	for _, target := range targets {
		key := s.targetKey(target)
		if seen[key] {
			return fmt.Errorf("duplicate target: %s", target.Label)
		}
		seen[key] = true
	}

	return nil
}

// targetKey generates a unique key for a target based on its attributes.
func (s *SetTargetImport) targetKey(target models.CoinSetTarget) string {
	parts := []string{strings.ToLower(target.Label)}

	if target.Year != nil {
		parts = append(parts, fmt.Sprintf("y%d", *target.Year))
	}
	if target.MintMark != nil {
		parts = append(parts, fmt.Sprintf("m%s", strings.ToLower(*target.MintMark)))
	}
	if target.Denomination != nil {
		parts = append(parts, fmt.Sprintf("d%s", strings.ToLower(*target.Denomination)))
	}

	return strings.Join(parts, "|")
}
