package user

type PassEncoder interface {
	Encode(string) string
}

type EncoderFunc func(s string) string

func (f EncoderFunc) Encode(s string) string {
	return f(s)
}
