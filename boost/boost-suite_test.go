package boost_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo ok
	. "github.com/onsi/gomega"    //nolint:revive // gomega ok
)

func TestAsync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Async Suite")
}
