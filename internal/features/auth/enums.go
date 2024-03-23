package auth

var (
	TokenTypeRefresh  int32 = 1
	TokenTypeRegister int32 = 2
	TokenTypeReset    int32 = 3
)

var (
	OtpTypeRegister int32 = 2
	OtpTypeReset    int32 = 3
)

var OtpToTokenType = map[int32]int32{
	OtpTypeRegister: TokenTypeRegister,
	OtpTypeReset:    TokenTypeReset,
}
