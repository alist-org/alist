package authninterface

type AuthnInterface interface {
	UpdateAuthn(userID uint, authn string) error
}
