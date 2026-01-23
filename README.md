This is a great choice. **Difftron** sounds like a high-performance utility—clean, technical, and fits perfectly into your "Tron" ecosystem.

Below is the rewritten project documentation. I have updated the branding, refined the architectural goals, and included the specialized Go logic for both the LCOV parsing and the Hunk mapping.

---

# Difftron

**Difftron** is a language-agnostic, AI-powered "Quality Gate" CLI. It ensures that new code changes are adequately tested by correlating `git diff` hunks with standard coverage reports (LCOV, Cobertura, etc.).

Unlike traditional coverage tools that report on the entire project, **Difftron** zooms in on the **deltas**, ensuring that every new line of code is held to a high standard of quality before it hits production.

## The Problem

In large repositories, a developer can introduce 100 lines of untested logic, and the overall project coverage might only drop by 0.01%. This makes it easy for "testing debt" to accumulate unnoticed.

## The Solution

Difftron parses the `git diff` to identify the specific line numbers modified in a PR. It then "intersects" this data with your coverage reports to identify exactly which new lines are untested. Finally, it uses Gemini AI to suggest the missing test cases.

---

## 🛠 Architecture

| Component | Responsibility |
| --- | --- |
| **Hunk Engine** | Parses `git diff` output to map "hunks" to absolute line numbers in the new file version. |
| **Coverage Engine** | Ingests LCOV/Cobertura files to build an in-memory map of `File -> Line -> HitCount`. |
| **Risk Engine** | Cross-references coverage with **Git Churn** (frequency of file changes) to flag high-risk gaps. |
| **Gemini Integration** | Provides "Agentic" test generation by sending uncovered hunks to the AI for boilerplate creation. |

---

## 💻 Implementation Highlights (Go)

### 1. The Hunk-to-Line Mapper

This core logic translates relative diff offsets into absolute line numbers that match your coverage report.

```go
// ParseGitDiff identifies absolute line numbers of added/modified code
func ParseGitDiff(diffOutput string) map[string]map[int]bool {
	changes := make(map[string]map[int]bool)
	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	var currentFile string
	var currentLine int

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			changes[currentFile] = make(map[int]bool)
			continue
		}

		// Header Example: @@ -10,5 +15,7 @@
		if strings.HasPrefix(line, "@@") {
			parts := strings.Split(line, " ")
			newFilePart := strings.Split(parts[2], ",") // e.g., +15
			startLine, _ := strconv.Atoi(newFilePart[0][1:])
			currentLine = startLine
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			changes[currentFile][currentLine] = true
			currentLine++
		} else if !strings.HasPrefix(line, "-") {
			currentLine++ // Context line
		}
	}
	return changes
}

```

### 2. The LCOV Coverage Parser

A high-performance scanner to ingest standardized coverage data.

```go
// ParseLCOV reads .info files into a File -> Line -> Hits map
func ParseLCOV(filePath string) (map[string]map[int]int, error) {
	file, _ := os.Open(filePath)
	defer file.Close()

	reports := make(map[string]map[int]int)
	var currentFile string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "SF:") {
			currentFile = line[3:]
			reports[currentFile] = make(map[int]int)
		} else if strings.HasPrefix(line, "DA:") {
			data := strings.Split(line[3:], ",")
			lineNum, _ := strconv.Atoi(data[0])
			hits, _ := strconv.Atoi(data[1])
			reports[currentFile][lineNum] = hits
		}
	}
	return reports, nil
}

```

---

## 🚀 Key Features

* **Universal Language Support**: If your language exports LCOV or Cobertura (Go, TS, Java, Python, C++, etc.), Difftron can analyze it.
* **Risk Heatmaps**: Uses Git Churn data to identify "Hot Spot" files. Low coverage in a high-churn file triggers a **CRITICAL** alert.
* **AI Test Generation**: Don't just report the gap—fix it. Difftron generates `_test.go` or `.test.ts` snippets for missing paths.
* **CI/CD Native**: Designed to run as a GitHub Action or GitLab CI Job, posting rich Markdown reports directly to your PR/MR.

---

## 📈 Roadmap

* [ ] **v0.1**: CLI Core (Diff + LCOV parsing).
* [ ] **v0.2**: Risk Scoring (Git Churn + Complexity).
* [ ] **v0.3**: GitHub Action/GitLab CI Commenter.
* [ ] **v0.4**: Gemini-powered Test Generation.

---
