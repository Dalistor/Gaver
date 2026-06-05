package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/Dalistor/gaver/templates"
)

type Data struct {
	Name      string
	Database  string
	CreatedAt string
}

func Generate(projectType, name, database string) error {
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("diretório %q já existe", name)
	}

	data := Data{
		Name:      name,
		Database:  database,
		CreatedAt: time.Now().Format("2006-01-02"),
	}

	return writeFromTemplates(projectType, name, data)
}

type ModuleData struct {
	ProjectName string
	ModuleName  string
}

func GenerateModule(moduleName string) error {
	m, err := manifest.Load()
	if err != nil {
		return err
	}

	outDir := filepath.Join("src", "modules", moduleName)
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		return fmt.Errorf("módulo %q já existe em %s", moduleName, outDir)
	}

	data := ModuleData{
		ProjectName: m.Name,
		ModuleName:  moduleName,
	}

	content, err := templates.FS.ReadFile("module/module.go.tmpl")
	if err != nil {
		return fmt.Errorf("template de módulo não encontrado: %w", err)
	}

	tmpl, err := template.New("module.go").Parse(string(content))
	if err != nil {
		return fmt.Errorf("template inválido: %w", err)
	}

	outPath := filepath.Join(outDir, "module.go")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func writeFromTemplates(templateRoot, projectDir string, data Data) error {
	return fs.WalkDir(templates.FS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, templateRoot+"/")
		if path == templateRoot {
			return nil
		}

		outPath := filepath.Join(projectDir, filepath.FromSlash(relPath))

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		content, err := templates.FS.ReadFile(path)
		if err != nil {
			return err
		}

		isTemplate := strings.HasSuffix(outPath, ".tmpl")
		if isTemplate {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		if isTemplate {
			tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
			if err != nil {
				return fmt.Errorf("template %s: %w", path, err)
			}
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()
			return tmpl.Execute(f, data)
		}

		return os.WriteFile(outPath, content, 0644)
	})
}
