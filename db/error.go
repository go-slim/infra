package db

const (
	ErrConfigRequired = Error("db: config is required")
	ErrDriverRequired = Error("db: driver is required")
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func (e Error) String() string {
	return string(e)
}
