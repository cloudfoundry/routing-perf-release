package plotgen_test

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"throughputramp/aggregator"
	"throughputramp/plotgen"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plotgen", func() {
	Describe("Generate", func() {
		It("returns a valid PNG plot", func() {
			points := testData(10)
			plotReader, err := plotgen.Generate(points)
			Expect(err).NotTo(HaveOccurred())
			defer plotReader.Close()

			plotBytes, err := ioutil.ReadAll(plotReader)
			Expect(err).NotTo(HaveOccurred())

			Expect(http.DetectContentType(plotBytes)).To(Equal("image/png"))
		})
	})
})

func testData(entry int) []aggregator.Point {
	var points aggregator.Report
	for i := 0; i <= entry; i++ {
		throughput := rand.Intn(100-50) + 50
		points = append(points, aggregator.Point{
			Throughput: float64(throughput),
			Latency:    time.Millisecond * time.Duration(i),
		})
	}
	return points
}
