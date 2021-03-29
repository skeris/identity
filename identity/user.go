package identity

type User struct {
	ID string `bson:"_id" json:"ID"`

	// TODO
	LastVerificationTime int64 `bson:"LastVerificationTime" json:"LastVerificationTime"`

	Identities        []IdentityData `bson:"Identities" json:"Identities"` // /name/identity/**
	Verifiers         []VerifierData `bson:"Verifiers" json:"Verifiers"`
	AuthFactorsNumber int            `bson:"AuthFactorsNumber" json:"AuthFactorsNumber"`

	Version int `bson:"Version" json:"Version"`
}

func (u *User) add(ver *VerifierData, idn *IdentityData) {
	if ver != nil {
		i := 0
		for i < len(u.Verifiers) {
			vd := u.Verifiers[i]
			if vd.Identity == ver.Identity && vd.Name == ver.Name {
				break
			}
			i++
		}
		if i == len(u.Verifiers) {
			u.Verifiers = append(u.Verifiers, *ver)
		} else {
			u.Verifiers[i] = *ver
		}
	}

	i := 0
	if len(u.Identities) != 0 {
		for i < len(u.Verifiers) {
			id := u.Identities[i]
			if id.Identity == idn.Identity && id.Name == idn.Name {
				break
			}
			i++
		}

		if i == len(u.Verifiers) {
			u.Identities = append(u.Identities, *idn)
		} else {
			u.Identities[i] = *idn
		}
	} else {
		u.Identities = append(u.Identities, *idn)
	}
}

type IdentityData struct {
	Name     string `bson:"Name" json:"Name"`
	Identity string `bson:"Identity" json:"Identity"`
}

type VerifierData struct {
	Name               string `bson:"Name" json:"Name"`
	Identity           string `bson:"Identity" json:"Identity"`
	AuthenticationData B      `bson:"AuthenticationData" json:"AuthenticationData"` // /identity/value
	AdditionalData     B      `bson:"AdditionalData" json:"AdditionalData"`
}

func (u *User) findVerifierData(verifierName, identity string) *VerifierData {
	for _, vd := range u.Verifiers {
		if vd.Name == verifierName {
			if identity != "" {
				if vd.Identity == identity {
					return &vd
				}
			} else {
				return &vd
			}
		}
	}
	return nil
}
