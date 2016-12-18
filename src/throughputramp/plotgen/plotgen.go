package plotgen

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"time"
)

func Generate(name string, csv, cpuCsv []byte, comparisonFile string) (io.Reader, error) {
	rscriptFile, err := createGeneratorRscript()
	if err != nil {
		return nil, err
	}
	defer cleanupGeneratorRscript(rscriptFile)

	csvFileName, err := loadFile(name, csv)
	if err != nil {
		return nil, err
	}
	defer cleanupTempFile(csvFileName)

	plotFileName := fmt.Sprintf("%s/plot.png%d", os.TempDir(), rand.Int())
	args := []string{rscriptFile, csvFileName, plotFileName}

	if comparisonFile != "" {
		args = append(args, "-comparefile", comparisonFile)
	}
	if cpuCsv != nil {
		tempCpuFileName := fmt.Sprintf("cpu-%s", time.Now().UTC().Format(time.RFC3339))
		cpuCsvFileName, err := loadFile(tempCpuFileName, cpuCsv)
		if err != nil {
			return nil, err
		}
		args = append(args, "-cpufile", cpuCsvFileName)
		defer cleanupTempFile(cpuCsvFileName)
	}
	cmd := exec.Command("Rscript", args...)
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Failed to run rscript command:\n%s\n%s", cmdOut, err.Error())
	}

	plotBuffer := bytes.NewBuffer(nil)

	plotFile, err := os.Open(plotFileName)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s: %s", plotFileName, err.Error())
	}
	defer cleanupTempFile(plotFile.Name())
	_, err = plotBuffer.ReadFrom(plotFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from %s: %s", plotFileName, err.Error())
	}

	return plotBuffer, nil
}

func loadFile(fileName string, csv []byte) (string, error) {
	csvPath := path.Join(os.TempDir(), fileName)
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return "", fmt.Errorf("Failed to create %s: %s", csvPath, err.Error())
	}
	_, err = csvFile.Write(csv)
	if err != nil {
		return "", fmt.Errorf("Failed to write to %s: %s", csvFile.Name(), err.Error())
	}
	return csvFile.Name(), nil
}

func cleanupTempFile(file string) {
	err := os.Remove(file)
	if err != nil {
		log.Printf("Failed to cleanup %s: %s", file, err.Error())
	}
}

func createGeneratorRscript() (string, error) {
	rscript, err := ioutil.TempFile("", "generate.r")
	if err != nil {
		return "", fmt.Errorf("Failed to create temp file generate.r: %s", err.Error())
	}

	defer func() {
		err = rscript.Close()
		if err != nil {
			log.Printf("Failed to close %s: %s", rscript.Name(), err.Error())
		}
	}()

	_, err = rscript.Write([]byte(generate_r))
	if err != nil {
		return "", fmt.Errorf("Failed to write to temp file generate.r: %s", err.Error())
	}
	return rscript.Name(), nil
}

func cleanupGeneratorRscript(rscriptFile string) error {
	err := os.Remove(rscriptFile)
	if err != nil {
		return fmt.Errorf("Failed to remove %s: %s", rscriptFile, err.Error())
	}
	return nil
}
