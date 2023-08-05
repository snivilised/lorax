package i18n

// CLIENT-TODO: Should be updated to use url of the implementing project,
// so should not be left as astrolib. (this should be set by auto-check)
const AstrolibSourceID = "github.com/snivilised/astrolib"

type astrolibTemplData struct{}

func (td astrolibTemplData) SourceID() string {
	return AstrolibSourceID
}
