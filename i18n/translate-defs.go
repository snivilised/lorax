package i18n

// CLIENT-TODO: Should be updated to use url of the implementing project,
// so should not be left as astrolib. (this should be set by auto-check)
const LoraxSourceID = "github.com/snivilised/lorax"

type loraxTemplData struct{}

func (td loraxTemplData) SourceID() string {
	return LoraxSourceID
}
