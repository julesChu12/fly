package constants

const (
	PasswordMinLength = 8
	PasswordMaxLength = 128

	UsernameMinLength = 3
	UsernameMaxLength = 50

	JWTIssuer = "custos-auth"
	JWTAccessTokenDuration = 15 // minutes
)