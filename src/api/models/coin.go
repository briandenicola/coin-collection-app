package models

import "time"

type Material string

const (
	MaterialGold     Material = "Gold"
	MaterialSilver   Material = "Silver"
	MaterialBronze   Material = "Bronze"
	MaterialCopper   Material = "Copper"
	MaterialElectrum Material = "Electrum"
	MaterialOther    Material = "Other"
)

type Category string

const (
	CategoryRoman     Category = "Roman"
	CategoryGreek     Category = "Greek"
	CategoryByzantine Category = "Byzantine"
	CategoryModern    Category = "Modern"
	CategoryOther     Category = "Other"
)

type ImageType string

const (
	ImageTypeObverse ImageType = "obverse"
	ImageTypeReverse ImageType = "reverse"
	ImageTypeDetail  ImageType = "detail"
	ImageTypeOther   ImageType = "other"
)

type Coin struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"not null" json:"name"`
	Category            Category   `gorm:"type:varchar(20);not null;default:'Other'" json:"category"`
	Denomination        string     `json:"denomination"`
	Ruler               string     `json:"ruler"`
	Era                 string     `json:"era"`
	Mint                string     `json:"mint"`
	Material            Material   `gorm:"type:varchar(20);default:'Other'" json:"material"`
	WeightGrams         *float64   `json:"weightGrams"`
	DiameterMm          *float64   `json:"diameterMm"`
	Grade               string     `json:"grade"`
	ObverseInscription  string     `json:"obverseInscription"`
	ReverseInscription  string     `json:"reverseInscription"`
	ObverseDescription  string     `json:"obverseDescription"`
	ReverseDescription  string     `json:"reverseDescription"`
	RarityRating        string     `json:"rarityRating"`
	PurchasePrice       *float64   `json:"purchasePrice"`
	CurrentValue        *float64   `json:"currentValue"`
	PurchaseDate        *time.Time `json:"purchaseDate"`
	PurchaseLocation    string     `json:"purchaseLocation"`
	Notes               string     `gorm:"type:text" json:"notes"`
	AIAnalysis          string     `gorm:"type:text;column:ai_analysis" json:"aiAnalysis"`
	ObverseAnalysis     string     `gorm:"type:text;column:obverse_analysis" json:"obverseAnalysis"`
	ReverseAnalysis     string     `gorm:"type:text;column:reverse_analysis" json:"reverseAnalysis"`
	ReferenceURL        string     `json:"referenceUrl"`
	ReferenceText       string     `json:"referenceText"`
	IsWishlist          bool       `gorm:"default:false" json:"isWishlist"`
	UserID              uint       `gorm:"not null" json:"userId"`
	User                User       `gorm:"foreignKey:UserID" json:"-"`
	Images              []CoinImage `gorm:"foreignKey:CoinID" json:"images"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type CoinImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CoinID    uint      `gorm:"not null" json:"coinId"`
	FilePath  string    `gorm:"not null" json:"filePath"`
	ImageType ImageType `gorm:"type:varchar(20);default:'other'" json:"imageType"`
	IsPrimary bool      `gorm:"default:false" json:"isPrimary"`
	CreatedAt time.Time `json:"createdAt"`
}
