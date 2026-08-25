// Command new-plugin scaffolds a new plugin package from the templates in
// hack/new-plugin/templates (see docs/plugin_development.md「插件脚手架」).
//
// Usage:
//
//	go run ./hack/new-plugin -name example -menu Example
//
// It generates <dir>/<name>/{registry.go,menu.go,<name>_test.go,README.md}
// (dir defaults to internal/plugins) from the plugin_name/ template directory,
// replacing the {{...}} placeholders. Generated Go files are gofmt-ed in
// place. Refuses to overwrite an existing target dir unless -force is given.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	pluginDirPlaceholder = "plugin_name"
	templatesDefault     = "hack/new-plugin/templates"
	outputDefault        = "internal/plugins"
	afterDefault         = "MainMenuStart"
)

var (
	flagName      = flag.String("name", "", "插件名：包名与目录名（小写 snake_case，必填）")
	flagMenu      = flag.String("menu", "", "菜单类型名前缀（CamelCase，如 Example -> ExampleMenu，必填）")
	flagKey       = flag.String("key", "", "菜单注册 key（默认 = -name）")
	flagTitle     = flag.String("title", "", "主菜单显示标题（默认 = -menu）")
	flagAfter     = flag.String("after", afterDefault, "主菜单 after 锚点（默认 MainMenuStart = 主菜单链首）")
	flagDir       = flag.String("dir", outputDefault, "输出插件目录（相对当前工作目录）")
	flagTemplates = flag.String("templates", templatesDefault, "模板目录（相对当前工作目录）")
	flagForce     = flag.Bool("force", false, "目标目录已存在时强制覆盖")
	flagSkipFmt   = flag.Bool("skip-fmt", false, "跳过 gofmt -w")
)

var (
	rePackageName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	reMenuName    = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
)

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "new-plugin: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	flag.Parse()

	if *flagName == "" || *flagMenu == "" {
		fail("usage: go run ./hack/new-plugin -name <插件名> -menu <菜单类型名> [-key <key>] [-title <标题>] [-after <锚点>]")
	}
	if !rePackageName.MatchString(*flagName) {
		fail("-name %q 非法：需小写 snake_case 包名（^[a-z][a-z0-9_]*$）", *flagName)
	}
	if !reMenuName.MatchString(*flagMenu) {
		fail("-menu %q 非法：需 CamelCase 类型前缀（^[A-Z][A-Za-z0-9]*$）", *flagMenu)
	}

	name := *flagName
	menu := *flagMenu
	key := *flagKey
	if key == "" {
		key = name
	}
	if !rePackageName.MatchString(key) {
		fail("-key %q 非法：需小写 snake_case（^[a-z][a-z0-9_]*$）", key)
	}
	title := *flagTitle
	if title == "" {
		title = menu
	}

	// after 表达式：MainMenuStart 渲染为 ui.MainMenuStart 常量，其余为带引号的 key。
	afterExpr := `"` + *flagAfter + `"`
	if *flagAfter == afterDefault {
		afterExpr = "ui.MainMenuStart"
	}

	// 内容占位符（带 {{}}）；文件名占位符是裸 plugin_name 前缀，单独替换。
	replacer := strings.NewReplacer(
		"{{plugin_name}}", name,
		"{{MenuType}}", menu,
		"{{menu_key}}", key,
		"{{menu_title}}", title,
		"{{menu_after}}", *flagAfter,
		"{{menu_after_expr}}", afterExpr,
	)
	fileReplacer := strings.NewReplacer(pluginDirPlaceholder, name)

	srcDir := filepath.Join(*flagTemplates, pluginDirPlaceholder)
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		fail("读取模板目录 %s: %v", srcDir, err)
	}

	targetDir := filepath.Join(*flagDir, name)
	if _, err := os.Stat(targetDir); err == nil && !*flagForce {
		fail("目标目录 %s 已存在（使用 -force 覆盖）", targetDir)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fail("创建 %s: %v", targetDir, err)
	}

	var generated []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".tmpl" {
			continue
		}
		tmplPath := filepath.Join(srcDir, entry.Name())
		content, err := os.ReadFile(tmplPath)
		if err != nil {
			fail("读取模板 %s: %v", tmplPath, err)
		}
		outName := fileReplacer.Replace(strings.TrimSuffix(entry.Name(), ".tmpl"))
		outPath := filepath.Join(targetDir, outName)
		rendered := replacer.Replace(string(content))
		if err := os.WriteFile(outPath, []byte(rendered), 0o644); err != nil {
			fail("写入 %s: %v", outPath, err)
		}
		generated = append(generated, outPath)
		if strings.HasSuffix(outName, ".go") && !*flagSkipFmt {
			if err := exec.Command("gofmt", "-w", outPath).Run(); err != nil {
				fmt.Fprintf(os.Stderr, "new-plugin: gofmt -w %s 失败（请手动格式化）: %v\n", outPath, err)
			}
		}
	}

	fmt.Printf("已生成 %d 个文件到 %s:\n", len(generated), targetDir)
	for _, p := range generated {
		fmt.Printf("  %s\n", p)
	}
	fmt.Printf("\n下一步：\n")
	fmt.Printf("  1. 在 internal/plugins/plugins.go 添加空导入：\n")
	fmt.Printf("       _ \"github.com/go-musicfox/go-musicfox/internal/plugins/%s\"\n", name)
	fmt.Printf("  2. 按需调整 registry.go 的主菜单锚点/标题与可选扩展\n")
	fmt.Printf("  3. gofmt -w %s && go build ./... && go test ./internal/plugins/%s/...\n", targetDir, name)
	fmt.Printf("  4. 用真实业务逻辑替换 menu.go 的静态项\n")
}
