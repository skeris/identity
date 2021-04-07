package identity_username

import (
	"errors"
	"github.com/skeris/identity/identity"
	"strings"
	"unicode"
)

var _ identity.Identity = new(Identity)

type Identity struct {
}

func New() *Identity {
	return &Identity{}
}

func (idn *Identity) Info() identity.IdentityInfo {

	return identity.IdentityInfo{
		Name: "username",
	}
}

func (i *Identity) NormalizeAndValidateIdentity(idn string) (idnNormalized string, err error) {
	for _, c := range idn {
		if !unicode.IsDigit(c) || !unicode.IsLetter(c) {
			return "", errors.New("invalid characters in username")
		}
	}
	return strings.ToLower(idn), nil

}
