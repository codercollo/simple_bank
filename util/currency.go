package util

//Supported currency codes
const (
	USD = "USD"
	EUR = "EUR"
	KSH = "Ksh"
)

//IsSupportedCurrency checks if currency is allowed
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, KSH:
		return true
	}
	return false
}
