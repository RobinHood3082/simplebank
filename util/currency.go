package util

const (
	USD = "USD"
	EUR = "EUR"
	GBP = "GBP"
	BDT = "BDT"
)

// IsCurrencySupported checks if the currency is supported
func IsCurrencySupported(currency string) bool {
	switch currency {
	case USD, EUR, GBP, BDT:
		return true
	}
	return false
}
