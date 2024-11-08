package constant

const (
	OtpLength       = 6
	SecretKeyLength = 32
)

var (
	LowerLetters = []rune("abcdefghijklmnopqrstuvwxyz")
	UpperLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Digits       = []rune("0123456789")
	SecretChars  = append(append(LowerLetters, UpperLetters...), Digits...)
)
