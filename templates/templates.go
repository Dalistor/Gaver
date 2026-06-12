// Package templates contém os arquivos de referência para repositórios Gaver-compatíveis.
// Os templates não são mais embutidos no binário — são baixados de repositórios Git
// registrados via `gaver repo add`. Veja a estrutura esperada de um repositório:
//
//	gaver-repo.json
//	projects/<type>/   ← templates de projeto
//	modules/           ← module.go.tmpl
package templates
