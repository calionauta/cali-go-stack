package auth

import (
	"github.com/a-h/templ"
	"github.com/calionauta/gogogo-fullstack-template/config"
	"github.com/calionauta/gogogo-fullstack-template/web/skins"
)

// SkinSelector renders the skin selector dropdown in the navbar.
// Wrapper avoids templ generate struggling with cross-package references.
func SkinSelector() templ.Component {
	return skins.SkinSelector(config.Get().Skin)
}
