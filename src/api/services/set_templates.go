package services

import (
	"fmt"

	"github.com/briandenicola/ancient-coins-api/models"
)

// SetTemplate represents a built-in set template for popular collecting series.
type SetTemplate struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Category    string              `json:"category"`
	Description string              `json:"description"`
	Targets     []SetTemplateTarget `json:"targets"`
	Version     int                 `json:"version"`
}

// SetTemplateTarget represents a target definition in a template.
type SetTemplateTarget struct {
	Label        string  `json:"label"`
	Year         *int    `json:"year"`
	MintMark     *string `json:"mintMark"`
	Denomination *string `json:"denomination"`
	Country      *string `json:"country"`
	Material     *string `json:"material"`
	SortOrder    int     `json:"sortOrder"`
}

// GetTemplates returns all available set templates.
func GetTemplates() []SetTemplate {
	return []SetTemplate{
		{
			ID:          "us-lincoln-wheat-1909-1958",
			Name:        "Lincoln Wheat Cents (1909-1958)",
			Category:    "US Cents",
			Description: "Complete set of Lincoln Wheat Cents from 1909 to 1958, including all mint marks",
			Version:     1,
			Targets:     generateLincolnWheatTargets(),
		},
		{
			ID:          "us-jefferson-nickel-1938-1964",
			Name:        "Jefferson Nickels (1938-1964)",
			Category:    "US Nickels",
			Description: "Complete set of Jefferson Nickels from 1938 to 1964, including war nickels",
			Version:     1,
			Targets:     generateJeffersonNickelTargets(),
		},
		{
			ID:          "us-washington-quarter-1932-1964",
			Name:        "Washington Quarters (1932-1964)",
			Category:    "US Quarters",
			Description: "Complete set of silver Washington Quarters from 1932 to 1964",
			Version:     1,
			Targets:     generateWashingtonQuarterTargets(),
		},
		{
			ID:          "us-mercury-dime-1916-1945",
			Name:        "Mercury Dimes (1916-1945)",
			Category:    "US Dimes",
			Description: "Complete set of Mercury (Winged Liberty Head) Dimes",
			Version:     1,
			Targets:     generateMercuryDimeTargets(),
		},
		{
			ID:          "us-state-quarters-1999-2008",
			Name:        "50 State Quarters (1999-2008)",
			Category:    "US Quarters",
			Description: "Complete set of 50 State Quarters from Philadelphia and Denver mints",
			Version:     1,
			Targets:     generateStateQuarterTargets(),
		},
	}
}

// GetTemplateByID retrieves a specific template by ID.
func GetTemplateByID(id string) *SetTemplate {
	templates := GetTemplates()
	for _, t := range templates {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

// CopyTemplateToCoinSetTargets converts template targets to CoinSetTarget models.
func CopyTemplateToCoinSetTargets(template *SetTemplate, setID uint) []models.CoinSetTarget {
	targets := make([]models.CoinSetTarget, 0, len(template.Targets))
	for _, t := range template.Targets {
		targets = append(targets, models.CoinSetTarget{
			SetID:        setID,
			Label:        t.Label,
			Year:         t.Year,
			MintMark:     t.MintMark,
			Denomination: t.Denomination,
			Country:      t.Country,
			Material:     t.Material,
			SortOrder:    t.SortOrder,
		})
	}
	return targets
}

// generateLincolnWheatTargets creates targets for Lincoln Wheat Cents.
func generateLincolnWheatTargets() []SetTemplateTarget {
	targets := []SetTemplateTarget{}
	sortOrder := 0

	// Notable years with mintmarks
	years := []struct {
		year  int
		mints []string
	}{
		{1909, []string{"", "S", "S VDB"}},
		{1910, []string{"", "S"}},
		{1911, []string{"", "D", "S"}},
		{1912, []string{"", "D", "S"}},
		{1913, []string{"", "D", "S"}},
		{1914, []string{"", "D", "S"}},
		{1915, []string{"", "D", "S"}},
		{1916, []string{"", "D", "S"}},
		{1917, []string{"", "D", "S"}},
		{1918, []string{"", "D", "S"}},
		{1919, []string{"", "D", "S"}},
		{1920, []string{"", "D", "S"}},
		{1921, []string{"", "S"}},
		{1922, []string{"D"}},
		{1923, []string{"", "S"}},
		{1924, []string{"", "D", "S"}},
		{1925, []string{"", "D", "S"}},
		{1926, []string{"", "D", "S"}},
		{1927, []string{"", "D", "S"}},
		{1928, []string{"", "D", "S"}},
		{1929, []string{"", "D", "S"}},
		{1930, []string{"", "D", "S"}},
		{1931, []string{"", "D", "S"}},
		{1932, []string{"", "D"}},
		{1933, []string{"", "D"}},
		{1934, []string{"", "D"}},
		{1935, []string{"", "D", "S"}},
		{1936, []string{"", "D", "S"}},
		{1937, []string{"", "D", "S"}},
		{1938, []string{"", "D", "S"}},
		{1939, []string{"", "D", "S"}},
		{1940, []string{"", "D", "S"}},
		{1941, []string{"", "D", "S"}},
		{1942, []string{"", "D", "S"}},
		{1943, []string{"", "D", "S"}},
		{1944, []string{"", "D", "S"}},
		{1945, []string{"", "D", "S"}},
		{1946, []string{"", "D", "S"}},
		{1947, []string{"", "D", "S"}},
		{1948, []string{"", "D", "S"}},
		{1949, []string{"", "D", "S"}},
		{1950, []string{"", "D", "S"}},
		{1951, []string{"", "D", "S"}},
		{1952, []string{"", "D", "S"}},
		{1953, []string{"", "D", "S"}},
		{1954, []string{"", "D", "S"}},
		{1955, []string{"", "D", "S"}},
		{1956, []string{"", "D"}},
		{1957, []string{"", "D"}},
		{1958, []string{"", "D"}},
	}

	denom := "Cent"
	country := "United States"
	material := "Bronze"

	for _, y := range years {
		for _, mint := range y.mints {
			label := ""
			var mintMark *string
			if mint == "" {
				label = fmt.Sprintf("%d", y.year)
				mintMark = nil
			} else {
				label = fmt.Sprintf("%d-%s", y.year, mint)
				mintMark = &mint
			}
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &y.year,
				MintMark:     mintMark,
				Denomination: &denom,
				Country:      &country,
				Material:     &material,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	return targets
}

// generateJeffersonNickelTargets creates targets for Jefferson Nickels.
func generateJeffersonNickelTargets() []SetTemplateTarget {
	targets := []SetTemplateTarget{}
	sortOrder := 0

	denom := "Five Cents"
	country := "United States"

	// 1938-1942: Regular composition
	regularMaterial := "Copper-Nickel"
	for year := 1938; year <= 1942; year++ {
		mints := []string{"", "D", "S"}
		if year == 1938 {
			mints = []string{"", "D"}
		}
		for _, mint := range mints {
			label := fmt.Sprintf("%d", year)
			var mintMark *string
			if mint != "" {
				label = fmt.Sprintf("%d-%s", year, mint)
				mintMark = &mint
			}
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &year,
				MintMark:     mintMark,
				Denomination: &denom,
				Country:      &country,
				Material:     &regularMaterial,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	// 1942-1945: War nickels (35% silver)
	warMaterial := "Silver"
	for year := 1942; year <= 1945; year++ {
		mints := []string{"P", "D", "S"}
		if year == 1942 {
			mints = []string{"P"}
		}
		for _, mint := range mints {
			label := fmt.Sprintf("%d-%s (War)", year, mint)
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &year,
				MintMark:     &mint,
				Denomination: &denom,
				Country:      &country,
				Material:     &warMaterial,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	// 1946-1964: Return to regular composition
	for year := 1946; year <= 1964; year++ {
		mints := []string{"", "D", "S"}
		if year >= 1955 {
			mints = []string{"", "D"}
		}
		for _, mint := range mints {
			label := fmt.Sprintf("%d", year)
			var mintMark *string
			if mint != "" {
				label = fmt.Sprintf("%d-%s", year, mint)
				mintMark = &mint
			}
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &year,
				MintMark:     mintMark,
				Denomination: &denom,
				Country:      &country,
				Material:     &regularMaterial,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	return targets
}

// generateWashingtonQuarterTargets creates targets for Washington Quarters.
func generateWashingtonQuarterTargets() []SetTemplateTarget {
	targets := []SetTemplateTarget{}
	sortOrder := 0

	denom := "Quarter Dollar"
	country := "United States"
	material := "Silver"

	// 1932-1964 silver quarters
	for year := 1932; year <= 1964; year++ {
		if year == 1933 {
			continue // No quarters minted in 1933
		}

		mints := []string{"", "D", "S"}
		if year >= 1956 {
			mints = []string{"", "D"}
		}

		for _, mint := range mints {
			label := fmt.Sprintf("%d", year)
			var mintMark *string
			if mint != "" {
				label = fmt.Sprintf("%d-%s", year, mint)
				mintMark = &mint
			}
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &year,
				MintMark:     mintMark,
				Denomination: &denom,
				Country:      &country,
				Material:     &material,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	return targets
}

// generateMercuryDimeTargets creates targets for Mercury Dimes.
func generateMercuryDimeTargets() []SetTemplateTarget {
	targets := []SetTemplateTarget{}
	sortOrder := 0

	denom := "Dime"
	country := "United States"
	material := "Silver"

	for year := 1916; year <= 1945; year++ {
		mints := []string{"", "D", "S"}
		if year == 1916 {
			mints = []string{"", "D", "S"}
		} else if year >= 1922 && year <= 1930 {
			mints = []string{"", "D", "S"}
		}

		for _, mint := range mints {
			label := fmt.Sprintf("%d", year)
			var mintMark *string
			if mint != "" {
				label = fmt.Sprintf("%d-%s", year, mint)
				mintMark = &mint
			}
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &year,
				MintMark:     mintMark,
				Denomination: &denom,
				Country:      &country,
				Material:     &material,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	return targets
}

// generateStateQuarterTargets creates targets for 50 State Quarters.
func generateStateQuarterTargets() []SetTemplateTarget {
	states := []struct {
		year  int
		state string
	}{
		{1999, "Delaware"}, {1999, "Pennsylvania"}, {1999, "New Jersey"}, {1999, "Georgia"}, {1999, "Connecticut"},
		{2000, "Massachusetts"}, {2000, "Maryland"}, {2000, "South Carolina"}, {2000, "New Hampshire"}, {2000, "Virginia"},
		{2001, "New York"}, {2001, "North Carolina"}, {2001, "Rhode Island"}, {2001, "Vermont"}, {2001, "Kentucky"},
		{2002, "Tennessee"}, {2002, "Ohio"}, {2002, "Louisiana"}, {2002, "Indiana"}, {2002, "Mississippi"},
		{2003, "Illinois"}, {2003, "Alabama"}, {2003, "Maine"}, {2003, "Missouri"}, {2003, "Arkansas"},
		{2004, "Michigan"}, {2004, "Florida"}, {2004, "Texas"}, {2004, "Iowa"}, {2004, "Wisconsin"},
		{2005, "California"}, {2005, "Minnesota"}, {2005, "Oregon"}, {2005, "Kansas"}, {2005, "West Virginia"},
		{2006, "Nevada"}, {2006, "Nebraska"}, {2006, "Colorado"}, {2006, "North Dakota"}, {2006, "South Dakota"},
		{2007, "Montana"}, {2007, "Washington"}, {2007, "Idaho"}, {2007, "Wyoming"}, {2007, "Utah"},
		{2008, "Oklahoma"}, {2008, "New Mexico"}, {2008, "Arizona"}, {2008, "Alaska"}, {2008, "Hawaii"},
	}

	targets := []SetTemplateTarget{}
	sortOrder := 0

	denom := "Quarter Dollar"
	country := "United States"
	material := "Copper-Nickel Clad"

	for _, s := range states {
		for _, mint := range []string{"P", "D"} {
			label := fmt.Sprintf("%d %s-%s", s.year, s.state, mint)
			targets = append(targets, SetTemplateTarget{
				Label:        label,
				Year:         &s.year,
				MintMark:     &mint,
				Denomination: &denom,
				Country:      &country,
				Material:     &material,
				SortOrder:    sortOrder,
			})
			sortOrder++
		}
	}

	return targets
}
