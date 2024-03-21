package i18n_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
)

func TestI18n(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "I18n Suite")
}
