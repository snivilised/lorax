package i18n_test

import (
	"github.com/snivilised/extendio/i18n"
)

const (
	GrafficoSourceID = "github.com/snivilised/graffico"
)

type GrafficoData struct{}

func (td GrafficoData) SourceID() string {
	return GrafficoSourceID
}

// 🧊 Pavement Graffiti Report

// PavementGraffitiReportTemplData
type PavementGraffitiReportTemplData struct {
	GrafficoData
	Primary string
}

func (td PavementGraffitiReportTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "pavement-graffiti-report.graffico.unit-test",
		Description: "Report of graffiti found on a pavement",
		Other:       "Found graffiti on pavement; primary colour: '{{.Primary}}'",
	}
}

// ☢️ Wrong Source Id

// WrongSourceIDTemplData
type WrongSourceIDTemplData struct {
	GrafficoData
}

func (td WrongSourceIDTemplData) SourceID() string {
	return "FOO-BAR"
}

func (td WrongSourceIDTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "wrong-source-id.graffico.unit-test",
		Description: "Incorrect Source Id for which doesn't match the one n the localizer",
		Other:       "Message with wrong id",
	}
}
