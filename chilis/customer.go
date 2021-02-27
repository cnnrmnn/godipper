package chilis

type Customer struct {
	Address   Address
	FirstName string
	LastName  string
	Phone     string
	Email     string
}

// validPhone returns true if the given string has 10 digit runes
func validPhone(phone string) bool {
	n := 0
	for _, digit := range phone {
		if digit < '0' || digit > '9' {
			return false
		}
		n++
	}
	if n != 10 {
		return false
	}
	return true
}

// validEmail returns true if the given string has at least one @ rune.
func validEmail(email string) bool {
	for _, c := range email {
		if c == '@' {
			return true
		}
	}
	return false
}

func (customer Customer) Valid() bool {
	return validPhone(customer.Phone) && validEmail(customer.Email)
}
