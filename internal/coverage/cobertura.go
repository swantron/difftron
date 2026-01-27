package coverage

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CoberturaCoverage represents the root element of a Cobertura XML file
type CoberturaCoverage struct {
	XMLName      xml.Name          `xml:"coverage"`
	LineRate     float64           `xml:"line-rate,attr"`
	BranchRate   float64           `xml:"branch-rate,attr"`
	LinesCovered int               `xml:"lines-covered,attr"`
	LinesValid   int               `xml:"lines-valid,attr"`
	Sources      CoberturaSources  `xml:"sources"`
	Packages     CoberturaPackages `xml:"packages"`
}

// CoberturaSources contains source paths
type CoberturaSources struct {
	Source []string `xml:"source"`
}

// CoberturaPackages contains package elements
type CoberturaPackages struct {
	Package []CoberturaPackage `xml:"package"`
}

// CoberturaPackage represents a package in Cobertura
type CoberturaPackage struct {
	Name       string           `xml:"name,attr"`
	LineRate   float64          `xml:"line-rate,attr"`
	BranchRate float64          `xml:"branch-rate,attr"`
	Classes    CoberturaClasses `xml:"classes"`
}

// CoberturaClasses contains class elements
type CoberturaClasses struct {
	Class []CoberturaClass `xml:"class"`
}

// CoberturaClass represents a class in Cobertura
type CoberturaClass struct {
	Name       string           `xml:"name,attr"`
	Filename   string           `xml:"filename,attr"`
	LineRate   float64          `xml:"line-rate,attr"`
	BranchRate float64          `xml:"branch-rate,attr"`
	Methods    CoberturaMethods `xml:"methods"`
	Lines      CoberturaLines   `xml:"lines"`
}

// CoberturaMethods contains method elements (optional, we mainly use lines)
type CoberturaMethods struct {
	Method []CoberturaMethod `xml:"method"`
}

// CoberturaMethod represents a method (optional, we mainly use lines)
type CoberturaMethod struct {
	Name       string         `xml:"name,attr"`
	Signature  string         `xml:"signature,attr"`
	LineRate   float64        `xml:"line-rate,attr"`
	BranchRate float64        `xml:"branch-rate,attr"`
	Lines      CoberturaLines `xml:"lines"`
}

// CoberturaLines contains line elements
type CoberturaLines struct {
	Line []CoberturaLine `xml:"line"`
}

// CoberturaLine represents a single line in Cobertura
type CoberturaLine struct {
	Number            int    `xml:"number,attr"`
	Hits              int    `xml:"hits,attr"`
	Branch            bool   `xml:"branch,attr"`
	ConditionCoverage string `xml:"condition-coverage,attr"`
}

// ParseCobertura parses a Cobertura XML format coverage file
// Returns a Report containing coverage data for all files
func ParseCobertura(filePath string) (*Report, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Cobertura file: %w", err)
	}
	defer file.Close()

	var cobertura CoberturaCoverage
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&cobertura); err != nil {
		return nil, fmt.Errorf("failed to parse Cobertura XML: %w", err)
	}

	report := &Report{
		FileCoverage: make(map[string]*CoverageData),
	}

	// Build a map of source paths for path resolution
	sourcePaths := make(map[string]bool)
	for _, source := range cobertura.Sources.Source {
		sourcePaths[source] = true
	}

	// Process all packages and classes
	for _, pkg := range cobertura.Packages.Package {
		for _, class := range pkg.Classes.Class {
			filePath := resolveFilePath(class.Filename, sourcePaths)

			// Normalize the file path
			filePath = NormalizePath(filePath)

			// Initialize coverage data for this file if not exists
			if report.FileCoverage[filePath] == nil {
				report.FileCoverage[filePath] = &CoverageData{
					LineHits: make(map[int]int),
				}
			}

			fileCoverage := report.FileCoverage[filePath]

			// Process lines from class level (preferred)
			for _, line := range class.Lines.Line {
				fileCoverage.LineHits[line.Number] = line.Hits
				fileCoverage.TotalLines++
				if line.Hits > 0 {
					fileCoverage.CoveredLines++
				}
			}

			// Also process lines from methods (some tools put lines here)
			for _, method := range class.Methods.Method {
				for _, line := range method.Lines.Line {
					// Only add if not already present (class-level takes precedence)
					if _, exists := fileCoverage.LineHits[line.Number]; !exists {
						fileCoverage.LineHits[line.Number] = line.Hits
						fileCoverage.TotalLines++
						if line.Hits > 0 {
							fileCoverage.CoveredLines++
						}
					}
				}
			}
		}
	}

	return report, nil
}

// resolveFilePath resolves a relative filename against source paths
func resolveFilePath(filename string, sourcePaths map[string]bool) string {
	// Try direct match first
	if sourcePaths[filename] {
		return filename
	}

	// Try each source path
	for sourcePath := range sourcePaths {
		fullPath := filepath.Join(sourcePath, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
		// Also try with normalized separators
		fullPath = strings.ReplaceAll(filepath.Join(sourcePath, filename), "\\", "/")
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// If no match found, return filename as-is (will be normalized later)
	return filename
}
