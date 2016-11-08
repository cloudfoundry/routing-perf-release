package plotgen_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlotgen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plotgen Suite")
}
