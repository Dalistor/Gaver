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
)

type Data struct {
	Name      string
	CreatedAt string
}

func Generate(repoDir, projectType, name string) error {
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("diretório %q já existe", name)
	}

	data := Data{
		Name:      name,
		CreatedAt: time.Now().Format("2006-01-02"),
	}

	templateRoot := filepath.Join(repoDir, "projects", projectType)
	fsys := os.DirFS(templateRoot)
	return writeFromTemplates(fsys, name, data)
}

type ModuleData struct {
	ProjectName string
	ModuleName  string
}

func GenerateModule(repoDir, moduleName string) error {
	m, err := manifest.Load()
	if err != nil {
		return err
	}

	outDir := filepath.Join("src", "modules", moduleName)
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		return fmt.Errorf("módulo %q já existe em %s", moduleName, outDir)
	}

	tmplPath := filepath.Join(repoDir, "modules", "module.go.tmpl")
	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("template de módulo não encontrado em %s: %w", tmplPath, err)
	}

	tmpl, err := template.New("module.go").Parse(string(content))
	if err != nil {
		return fmt.Errorf("template inválido: %w", err)
	}

	data := ModuleData{
		ProjectName: m.Name,
		ModuleName:  moduleName,
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

func writeFromTemplates(fsys fs.FS, projectDir string, data Data) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		outPath := filepath.Join(projectDir, filepath.FromSlash(path))

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		content, err := fs.ReadFile(fsys, path)
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
