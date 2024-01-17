package watch

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	ui "tonysoft.com/gasp"
	"tonysoft.com/gasp/gowatch/console"
)

const (
	gaspPackage        = "tonysoft.com/gasp"
	gowatchPackage     = "tonysoft.com/gasp/gowatch"
	defaultControlType = "line"
)

type Watch struct {
	Id            string
	Path          string
	LineNumber    int
	Expression    string
	ControlType   string
	ControlConfig map[string]interface{}
}

func Start() error {
	sessionId, err := newSessionGuid()
	if err != nil {
		return err
	}

	err = placeCurrentDirMarker(sessionId)
	if err != nil {
		return err
	}

	sessionDir, err := createSessionDir(sessionId)
	if err != nil {
		return err
	}

	err = copyTargetProject(sessionDir)
	if err != nil {
		return err
	}

	err = copyCurrentDirMarker(sessionDir, sessionId)
	if err != nil {
		return err
	}

	err = removeCurrentDirMarker(sessionId)
	if err != nil {
		return err
	}

	err = formatCode(sessionDir, sessionId)
	if err != nil {
		return err
	}

	watches, err := findWatches(sessionDir)
	if err != nil {
		return err
	}

	err = injectWatchCode(watches)
	if err != nil {
		return err
	}

	err = injectImportCode(watches)
	if err != nil {
		return err
	}

	err = injectInitCode(sessionDir, getInitCode(watches))
	if err != nil {
		return err
	}

	err = injectModCode(sessionDir)
	if err != nil {
		return err
	}

	err = formatCode(sessionDir, sessionId)
	if err != nil {
		return err
	}

	return runCode(sessionDir, sessionId)
}

func newSessionGuid() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func createSessionDir(sessionId string) (string, error) {
	path := fmt.Sprintf("/tmp/.gowatch/session/%s", sessionId)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func copyTargetProject(sessionDir string) error {
	rootDir := console.Args.RootDir
	if rootDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		rootDir = wd
	}

	return exec.Command("/bin/sh", "-c", fmt.Sprintf("cp -r '%s'/* '%s'", rootDir, sessionDir)).Run()
}

func placeCurrentDirMarker(seshId string) error {
	markerContent := []byte(seshId)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	markerFilename := wd + "/.gowatch-" + seshId
	return os.WriteFile(markerFilename, markerContent, 0644)
}

func copyCurrentDirMarker(sessionDir, sessionId string) error {
	return exec.Command("/bin/sh", "-c", fmt.Sprintf("cp -r '.gowatch-%s' '%s'", sessionId, sessionDir)).Run()
}

func removeCurrentDirMarker(sessionId string) error {
	return exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -rf '.gowatch-%s'", sessionId)).Run()
}

func getWorkingDir(sessionDir, sessionId string) (string, error) {
	wd := ""
	err := filepath.Walk(sessionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), ".gowatch-"+sessionId) {
			wd = strings.Replace(path, info.Name(), "", 1)
		}

		return nil
	})
	return wd, err
}

func formatCode(sessionDir, sessionId string) error {
	wd, err := getWorkingDir(sessionDir, sessionId)
	if err != nil {
		return err
	}

	goMod, err := os.Open(wd + "/go.mod")
	if err != nil {
		return err
	}
	defer goMod.Close()

	module := ""
	reader := bufio.NewReader(goMod)
	for {
		line, err := reader.ReadString(10)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		lineParts := strings.Split(line, " ")
		if len(lineParts) == 2 && lineParts[0] == "module" {
			module = lineParts[1]
			break
		}
	}

	if module == "" {
		return errors.New("module not found")
	}

	module = module[:strings.LastIndex(module, "/")+1] + "..."

	cmd := exec.Command("go", "clean")
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("go", "fmt", module)
	cmd.Dir = wd
	return cmd.Run()
}

func findWatches(sessionDir string) ([]*Watch, error) {
	watches := make([]*Watch, 0)

	err := filepath.Walk(sessionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		lineNumber := 0
		for {
			lineNumber++
			line, err := reader.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			line = strings.TrimSpace(line)

			if !strings.HasPrefix(line, "//gowatch ") {
				continue
			}

			watch := Watch{
				Path:       path,
				LineNumber: lineNumber,
			}

			line = strings.TrimSpace(strings.TrimPrefix(line, "//gowatch "))

			if strings.HasPrefix(line, "{") {
				lastBraceIdx := strings.LastIndex(line, "}")
				watch.Expression = line[1:lastBraceIdx]
				line = line[lastBraceIdx+1:]
			} else {
				watch.Expression = strings.Split(line, " ")[0]
				watch.Id = watch.Expression
				line = line[len(watch.Expression):]
			}

			lineParts := strings.Split(line, " ")
			watch.ControlConfig = make(map[string]interface{})
			for _, part := range lineParts {
				kv := strings.Split(part, "=")
				if len(kv) != 2 {
					continue
				}

				key := kv[0]
				val := kv[1]

				if key == "id" {
					watch.Id = val
					continue
				} else if key == "type" {
					watch.ControlType = val
					continue
				}

				watch.ControlConfig[key] = val
			}

			if watch.ControlType == "" {
				watch.ControlType = defaultControlType
			}

			watches = append(watches, &watch)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return watches, nil
}

func injectWatchCode(watches []*Watch) error {
	for _, watch := range watches {
		file, err := os.Open(watch.Path)
		if err != nil {
			return err
		}

		var sb strings.Builder

		reader := bufio.NewReader(file)
		lineNumber := 0
		var readErr error

		for {
			lineNumber++
			line, err := reader.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				readErr = err
				break
			}

			if lineNumber == watch.LineNumber {
				tabCount := len(line) - len(strings.TrimLeft(line, "\t"))
				for i := 0; i < tabCount; i++ {
					sb.WriteString("\t")
				}

				sb.WriteString(getWatchCode(watch) + "\n")
			} else {
				sb.WriteString(line)
			}
		}

		err = file.Close()
		if readErr != nil {
			return readErr
		}
		if err != nil {
			return err
		}

		file, err = os.Create(watch.Path)
		if err != nil {
			return err
		}

		_, err = file.WriteString(sb.String())
		if err != nil {
			_ = file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func getWatchCode(watch *Watch) string {
	switch watch.ControlType {
	case "line":
		return fmt.Sprintf("_gowatch.UpdateLineChart(\"%s\", (%s))", watch.Id, watch.Expression)
	}
	return "UNKNOWN CONTROL TYPE: " + watch.ControlType
}

func injectImportCode(watches []*Watch) error {
	importAdded := make(map[string]bool)

	for _, watch := range watches {
		file, err := os.Open(watch.Path)
		if err != nil {
			return err
		}

		var sb strings.Builder

		reader := bufio.NewReader(file)
		var readErr error

		for {
			line, err := reader.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				readErr = err
				break
			}

			if strings.HasPrefix(line, "package ") {
				sb.WriteString(line)
				if added, ok := importAdded[watch.Path]; !ok || !added {
					sb.WriteString(fmt.Sprintf("\nimport _gowatch \"%s/ui\"\n", gowatchPackage))
					importAdded[watch.Path] = true
				}
				continue
			}

			sb.WriteString(line)
		}

		err = file.Close()
		if readErr != nil {
			return readErr
		}
		if err != nil {
			return err
		}

		file, err = os.Create(watch.Path)
		if err != nil {
			return err
		}

		_, err = file.WriteString(sb.String())
		if err != nil {
			_ = file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func injectInitCode(sessionDir string, initCode string) error {
	err := filepath.Walk(sessionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var sb strings.Builder
		isMainPackage := false
		gaspImported := false
		gowatchImported := false
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if strings.HasPrefix(line, "import _gasp") {
				gaspImported = true
			}
			if strings.HasPrefix(line, "import _gowatch") {
				gowatchImported = true
			}

			if strings.TrimSpace(line) == "package main" {
				isMainPackage = true
				sb.WriteString(line)
				continue
			}

			if !isMainPackage || strings.TrimSpace(line) != "func main() {" {
				sb.WriteString(line)
				continue
			}

			if !gaspImported {
				sb.WriteString(fmt.Sprintf("\nimport _gasp \"%s\"\n", gaspPackage))
			}
			if !gowatchImported {
				sb.WriteString(fmt.Sprintf("\nimport _gowatch \"%s/ui\"\n", gowatchPackage))
			}

			sb.WriteString(line)
			sb.WriteString(initCode)
		}

		file, err = os.Create(path)
		if err != nil {
			return err
		}

		_, err = file.WriteString(sb.String())
		if err != nil {
			_ = file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func getInitCode(watches []*Watch) string {
	var sb strings.Builder
	lines := make(map[string]*ui.LineChartState)

	for _, watch := range watches {
		switch watch.ControlType {
		case "line":
			line := lines[watch.Id]
			if line == nil {
				line = &ui.LineChartState{}
				line.Id = watch.Id
				line.Lines = make([]*ui.LineState, 1)
				line.Lines[0] = &ui.LineState{
					Name: "line0",
				}
				lines[watch.Id] = line
			}
			for k, v := range watch.ControlConfig {
				numVal, _ := strconv.Atoi(v.(string))

				switch k {
				case "text":
					line.Text = v.(string)
				case "width":
					line.Width = numVal
				case "height":
					line.Height = numVal
				case "thickness":
					line.Lines[0].Thickness = numVal
				case "color":
					line.Lines[0].Color = v.(string)
				}
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\t_gowatch.Init(\"%s\")\n", console.Args.Socket))

	lineIndex := 0
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("\t_gline%d := _gasp.LineChartState{}\n", lineIndex))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Id = \"%s\"\n", lineIndex, line.Id))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Text = \"%s\"\n", lineIndex, line.Text))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Width = %d\n", lineIndex, line.Width))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Height = %d\n", lineIndex, line.Height))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Lines = make([]*_gasp.LineState, 1)\n", lineIndex))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Lines[0] = &_gasp.LineState{}\n", lineIndex))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Lines[0].Thickness = %d\n", lineIndex, line.Lines[0].Thickness))
		sb.WriteString(fmt.Sprintf("\t_gline%d.Lines[0].Color = \"%s\"\n", lineIndex, line.Lines[0].Color))
		sb.WriteString(fmt.Sprintf("\t_gowatch.AddLineChart(_gline%d)\n", lineIndex))

		lineIndex++
	}

	sb.WriteString("\t_gowatch.Start()\n")
	sb.WriteString("\tdefer _gowatch.Stop()\n\n")

	return sb.String()
}

func injectModCode(sessionDir string) error {
	gaspPackageOverride := os.Getenv("GASPHOME")
	if gaspPackageOverride == "" {
		return nil
	}
	if strings.HasSuffix(gaspPackageOverride, "/") {
		gaspPackageOverride = gaspPackageOverride[0 : len(gaspPackageOverride)-1]
	}

	err := filepath.Walk(sessionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() != "go.mod" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var sb strings.Builder
		reader := bufio.NewReader(file)
		replaceAdded := false
		for {
			line, err := reader.ReadString(10)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			sb.WriteString(line)

			line = strings.TrimSpace(line)
			if !replaceAdded && strings.HasPrefix(line, "go ") {
				sb.WriteString(fmt.Sprintf("\nreplace %s => %s\n", gaspPackage, gaspPackageOverride))
				sb.WriteString(fmt.Sprintf("\nreplace %s => %s\n", gowatchPackage, gaspPackageOverride+"/gowatch"))
				replaceAdded = true
				continue
			}
		}

		file, err = os.Create(path)
		if err != nil {
			return err
		}

		_, err = file.WriteString(sb.String())
		if err != nil {
			_ = file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func runCode(sessionDir, sessionId string) error {
	wd, err := getWorkingDir(sessionDir, sessionId)

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = wd
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	buff := bufio.NewScanner(os.Stdin)
	go func() {
		for buff.Scan() {
			input := buff.Text()
			_, err := io.WriteString(stdin, input)
			if err != nil {
				panic(err)
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}
