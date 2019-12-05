package utils

type Size int64

const (
	minSize Size = -1 << 63
	maxSize Size = 1<<63 - 1
)

const (
	Byte     Size = 1
	KiloByte      = Byte << 10
	MegaByte      = KiloByte << 10
	Gigabyte      = MegaByte << 10
	TeraByte      = Gigabyte << 10
	PetaByte      = TeraByte << 10
	ExaByte       = PetaByte << 10
)

// Byte returns the size as a Byte count
func (s Size) Byte() int64 { return int64(s) }

// KiloByte returns the size as an integer KiloByte count.
func (s Size) KiloByte() int64 { return int64(s) >> 10 }

// MegaByte returns the size as an integer MegaByte count.
func (s Size) MegaByte() int64 { return int64(s) >> 20 }

// Gigabyte returns the size as an integer Gigabyte count.
func (s Size) Gigabyte() int64 { return int64(s) >> 30 }

// TeraByte returns the size as an integer TeraByte count.
func (s Size) TeraByte() int64 { return int64(s) >> 40 }

// PetaByte returns the size as an integer PetaByte count.
func (s Size) PetaByte() int64 { return int64(s) >> 50 }

// ExaByte returns the size as an integer ExaByte count.
func (s Size) ExaByte() int64 { return int64(s) >> 60 }
