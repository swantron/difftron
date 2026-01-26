package coverage

import (
	"os"
	"testing"
)

func TestParseCobertura(t *testing.T) {
	// Create a test Cobertura XML file
	coberturaContent := `<?xml version="1.0"?>
<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">
<coverage line-rate="0.5" branch-rate="0.5" lines-covered="3" lines-valid="6" branches-covered="0" branches-valid="0" complexity="0" version="0.1" timestamp="1234567890">
  <sources>
    <source>/path/to/source</source>
  </sources>
  <packages>
    <package name="com.example" line-rate="0.5" branch-rate="0.5">
      <classes>
        <class name="MyClass" filename="src/com/example/MyClass.java" line-rate="0.5" branch-rate="0.5">
          <methods>
            <method name="method1" signature="()V" line-rate="1.0" branch-rate="1.0">
              <lines>
                <line number="10" hits="5"/>
                <line number="11" hits="3"/>
              </lines>
            </method>
            <method name="method2" signature="()V" line-rate="0.0" branch-rate="0.0">
              <lines>
                <line number="15" hits="0"/>
              </lines>
            </method>
          </methods>
          <lines>
            <line number="10" hits="5"/>
            <line number="11" hits="3"/>
            <line number="15" hits="0"/>
          </lines>
        </class>
        <class name="AnotherClass" filename="src/com/example/AnotherClass.java" line-rate="1.0" branch-rate="1.0">
          <lines>
            <line number="5" hits="10"/>
            <line number="6" hits="8"/>
            <line number="7" hits="12"/>
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>
`

	tmpfile, err := os.CreateTemp("", "test-*.xml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(coberturaContent)); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpfile.Close()

	// Parse the Cobertura file
	report, err := ParseCobertura(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check file coverage
	file1Coverage := report.GetCoverageForFile("src/com/example/MyClass.java")
	if file1Coverage == nil {
		// Try normalized path
		file1Coverage = report.GetCoverageForFile("src/com/example/MyClass.java")
	}
	if file1Coverage == nil {
		t.Fatal("expected coverage data for MyClass.java")
	}

	if file1Coverage.TotalLines != 3 {
		t.Errorf("expected 3 total lines for MyClass.java, got %d", file1Coverage.TotalLines)
	}

	if file1Coverage.CoveredLines != 2 {
		t.Errorf("expected 2 covered lines for MyClass.java, got %d", file1Coverage.CoveredLines)
	}

	// Check specific line coverage
	if report.GetCoverageForLine("src/com/example/MyClass.java", 10) != 5 {
		t.Errorf("expected line 10 to have 5 hits, got %d", report.GetCoverageForLine("src/com/example/MyClass.java", 10))
	}

	if report.GetCoverageForLine("src/com/example/MyClass.java", 15) != 0 {
		t.Errorf("expected line 15 to have 0 hits, got %d", report.GetCoverageForLine("src/com/example/MyClass.java", 15))
	}

	if report.IsLineCovered("src/com/example/MyClass.java", 10) != true {
		t.Error("expected line 10 to be covered")
	}

	if report.IsLineCovered("src/com/example/MyClass.java", 15) != false {
		t.Error("expected line 15 to not be covered")
	}

	// Check second file
	file2Coverage := report.GetCoverageForFile("src/com/example/AnotherClass.java")
	if file2Coverage == nil {
		t.Fatal("expected coverage data for AnotherClass.java")
	}

	if file2Coverage.TotalLines != 3 {
		t.Errorf("expected 3 total lines for AnotherClass.java, got %d", file2Coverage.TotalLines)
	}

	if file2Coverage.CoveredLines != 3 {
		t.Errorf("expected 3 covered lines for AnotherClass.java, got %d", file2Coverage.CoveredLines)
	}
}

func TestParseCobertura_EmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-empty-*.xml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	report, err := ParseCobertura(tmpfile.Name())
	if err == nil {
		t.Error("expected error for empty Cobertura file")
	}
	if report != nil {
		t.Error("expected nil report for empty file")
	}
}

func TestParseCobertura_InvalidFile(t *testing.T) {
	_, err := ParseCobertura("/nonexistent/file.xml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParseCobertura_InvalidXML(t *testing.T) {
	invalidXML := `<?xml version="1.0"?>
<coverage>
  <invalid>content</invalid>
</coverage>
`

	tmpfile, err := os.CreateTemp("", "test-invalid-*.xml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(invalidXML)); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpfile.Close()

	_, err = ParseCobertura(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid Cobertura XML")
	}
}
