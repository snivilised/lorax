package i18n

const LoraxSourceID = "github.com/snivilised/lorax"

type loraxTemplData struct{}

func (td loraxTemplData) SourceID() string {
	return LoraxSourceID
}
