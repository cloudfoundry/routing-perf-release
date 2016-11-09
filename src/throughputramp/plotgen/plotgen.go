package plotgen

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"throughputramp/aggregator"
)

func Generate(r aggregator.Report) (io.ReadCloser, error) {
	rscriptFile, err := createGeneratorRscript()
	if err != nil {
		return nil, err
	}

	csvFile, err := ioutil.TempFile("", "data.csv")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp file data.csv: %s", err.Error())
	}

	_, err = csvFile.Write([]byte(r.GenerateCSV()))
	if err != nil {
		return nil, fmt.Errorf("Failed to write to temp file data.csv: %s", err.Error())
	}
	err = csvFile.Close()
	if err != nil {
		return nil, fmt.Errorf("Failed to close temp file data.csv: %s", err.Error())
	}

	plotFileName := fmt.Sprintf("%s/plot.png%d", os.TempDir(), rand.Int())

	cmd := exec.Command("Rscript", rscriptFile, csvFile.Name(), plotFileName)

	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Failed to run rscript command:\n%s\n%s", cmdOut, err.Error())
	}

	err = os.Remove(csvFile.Name())
	if err != nil {
		return nil, fmt.Errorf("Failed to remove temp file data.csv: %s", err.Error())
	}

	err = os.Remove(rscriptFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to remove temp file generate.r: %s", err.Error())
	}

	plotFile, err := os.Open(plotFileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open temp file plot.png: %s", err.Error())
	}
	return plotFile, nil
}

func createGeneratorRscript() (string, error) {
	rscript, err := ioutil.TempFile("", "generate.r")
	if err != nil {
		return "", fmt.Errorf("Failed to create temp file generate.r: %s", err.Error())
	}

	_, err = rscript.Write([]byte(generate_r))
	if err != nil {
		return "", fmt.Errorf("Failed to write to temp file generate.r: %s", err.Error())
	}
	err = rscript.Close()
	if err != nil {
		return "", fmt.Errorf("Failed to close temp file generate.r: %s", err.Error())
	}
	return rscript.Name(), nil
}
