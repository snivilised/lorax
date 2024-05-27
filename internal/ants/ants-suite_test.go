package ants_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // ok
	. "github.com/onsi/gomega"    //nolint:revive // ok
)

func TestAnts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ants Suite")
}
