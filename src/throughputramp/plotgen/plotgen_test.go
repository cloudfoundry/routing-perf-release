package plotgen_test

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"throughputramp/aggregator"
	"throughputramp/plotgen"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const comparisonCSV = `
throughput,latency
100, 0.01
200, 0.02
300, 0.03
400, 0.04
500, 0.05
600, 0.06
`

var _ = Describe("Plotgen", func() {
	Describe("Generate", func() {
		It("returns a valid PNG plot", func() {
			points := testData(10)
			plotReader, err := plotgen.Generate("test", points.GenerateCSV(), "")
			Expect(err).NotTo(HaveOccurred())

			plotBytes, err := ioutil.ReadAll(plotReader)
			Expect(err).NotTo(HaveOccurred())

			Expect(http.DetectContentType(plotBytes)).To(Equal("image/png"))
		})

		Context("with a valid comparison file", func() {
			var comparisonFilePath string
			BeforeEach(func() {
				comparisonFile, err := ioutil.TempFile("", "comparison.csv")
				Expect(err).ToNot(HaveOccurred())

				comparisonFile.WriteString(comparisonCSV)
				comparisonFilePath = comparisonFile.Name()
			})

			AfterEach(func() {
				err := os.Remove(comparisonFilePath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns a valid PNG plot", func() {
				points := testData(10)
				plotReader, err := plotgen.Generate("test", points.GenerateCSV(), comparisonFilePath)
				Expect(err).NotTo(HaveOccurred())

				plotBytes, err := ioutil.ReadAll(plotReader)
				Expect(err).NotTo(HaveOccurred())

				Expect(http.DetectContentType(plotBytes)).To(Equal("image/png"))
			})
		})

		Context("with an invalid comparison file path", func() {
			It("returns an error", func() {
				points := testData(10)
				_, err := plotgen.Generate("test", points.GenerateCSV(), "/does/not/exist")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func testData(entry int) aggregator.Report {
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
