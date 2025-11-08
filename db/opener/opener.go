package opener

import (
	"fmt"

	"gorm.io/gorm"
)

var openers map[string]Opener

func init() {
	openers = make(map[string]Opener)
}

type Opener interface {
	Name() string
	Open(map[string]any) (*gorm.DB, error)
}

func Register(opener Opener) {
	name := opener.Name()
	if _, ok := openers[name]; ok {
		panic(fmt.Sprintf("opener %s already registered", name))
	}

	openers[name] = opener
}

func Open(name string, config map[string]any) (*gorm.DB, error) {
	opener, ok := openers[name]
	if !ok {
		return nil, fmt.Errorf("opener %s not found", name)
	}
	return opener.Open(config)
}
