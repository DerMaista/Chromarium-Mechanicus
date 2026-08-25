package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kingpin/v2"
)

type Config struct {
	ThemesDir Path `json:"themesDir"`
}

func findConfigFile(file string) string {
	file, err := xdg.ConfigFile("chromarium-mechanicus/" + file)
	if err != nil {
		log.Fatal(file, "doesnt exist => save at", xdg.ConfigDirs, "chromarium-mechanicus/"+file)
		return ""
	}

	return file
}

func loadConfig(path string, mustExist bool) Config {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		if mustExist {
			log.Fatalf("error opening config file %s: %s", path, err)
		}
		slog.Debug("no config file at " + path + ", using defaults")
		return config
	}
	defer file.Close()

	ReadJson(file, &config)
	if config.ThemesDir != "" {
		config.ThemesDir.Resolve(filepath.Dir(path))
	}

	return config
}

func findConfigDir(dir string) string {
	dir, err := xdg.ConfigFile("chromarium-mechanicus/" + dir)
	if err != nil {
		log.Fatal(dir, "doesnt exist => save at", xdg.ConfigDirs, "chromarium-mechanicus/"+dir)
		return ""
	}
	return dir
}

var (
	configFileSetByUser bool
	themesDirSetByUser  bool
)

var (
	debug        = kingpin.Flag("debug", "Enable debug mode.").Short('v').Bool()
	configFile   = kingpin.Flag("config", "alternative config file to use instead of xdgConfigHome/chromarium-mechanicus/config.json").Short('c').IsSetByUser(&configFileSetByUser).Default(findConfigFile("config.json")).String()
	templateFile = kingpin.Flag("template", "alternative template file to use instead of xdgConfigHome/chromarium-mechanicus/template.json").Short('t').Default(findConfigFile("template.json")).File()
	templatesDir = kingpin.Flag("templatesDir", "alternative directory to resolve relatives paths inside the template.json file to").Default(findConfigFile("")).String()
	themesDir    = kingpin.Flag("themesDir", "alternative directory to search for themes in").IsSetByUser(&themesDirSetByUser).Default(findConfigDir("themes")).String()
	themeName    = kingpin.Arg("theme", "Theme to use.").Required().String()
)

type Color string

type Path string

func (p *Path) Resolve(baseDir string) Path {
	// Expand first: "$HOME/foo" is absolute once expanded and must not be
	// joined onto baseDir.
	p.ResolveEnv()
	if !filepath.IsAbs(string(*p)) {
		*p = Path(filepath.Join(baseDir, string(*p)))
	}
	*p = Path(filepath.Clean(string(*p)))
	return *p
}

func (p *Path) ResolveEnv() Path {
	*p = Path(os.ExpandEnv(string(*p)))
	return *p
}

func (p Path) Open() *os.File {
	file, err := os.Open(string(p))
	if err != nil {
		log.Fatal("error opening file " + string(p) + ": " + err.Error())
	}
	return file
}

func ReadFile(file *os.File) []byte {
	byteFile, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("error reading file: %s", err)
	}
	return byteFile
}

func ReadJson(file *os.File, v any) {
	byteFile := ReadFile(file)

	err := json.Unmarshal(byteFile, v)
	if err != nil {
		log.Fatalf("error unmarshaling file: %s", err)
	}
}

type TemplateFile []byte

func ReadTemplate(template *os.File) TemplateFile {
	return TemplateFile(ReadFile(template))
}
func (f *TemplateFile) Replace(data any) TemplateFile {
	t, err := template.New("template").Parse(string(*f))
	if err != nil {
		log.Fatalf("error replacing %v: %s", f, err)
		return nil
	}
	var b bytes.Buffer
	err = t.Execute(&b, data)
	*f = TemplateFile(b.String())

	return *f
}

type Cmd string

func (c Cmd) Run() ([]byte, error) {
	cmd := exec.Command("sh", "-c", string(c))

	return cmd.CombinedOutput()
}
func (c *Cmd) Replace(data any) Cmd {
	t, err := template.New("cmd").Parse(string(*c))
	if err != nil {
		log.Fatalf("error replacing %v: %s", c, err)
		return *c
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)
	*c = Cmd(b.String())

	return *c
}

type ThemeMode string

const (
	Light ThemeMode = "light"
	Dark  ThemeMode = "dark"
)

type Theme struct {
	Colors    Colors    `json:"colors"`
	Wallpaper Path      `json:"wallpaper"` // Absolute path to wallpaper
	Mode      ThemeMode `json:"mode"`      // "light" or "dark"
}

type Colors struct {
	// Backgrounds
	Background      Color `json:"background"`
	BackgroundOn    Color `json:"background_on"`
	BackgroundMuted Color `json:"background_muted"`

	Surface      Color `json:"surface"`
	SurfaceOn    Color `json:"surface_on"`
	SurfaceMuted Color `json:"surface_muted"`

	// Brand colors
	Primary      Color `json:"primary"`
	PrimaryOn    Color `json:"primary_on"`
	PrimaryMuted Color `json:"primary_muted"`

	Secondary      Color `json:"secondary"`
	SecondaryOn    Color `json:"secondary_on"`
	SecondaryMuted Color `json:"secondary_muted"`

	Tertiary      Color `json:"tertiary"`
	TertiaryOn    Color `json:"tertiary_on"`
	TertiaryMuted Color `json:"tertiary_muted"`

	// Status colors
	Success      Color `json:"success"`
	SuccessOn    Color `json:"success_on"`
	SuccessMuted Color `json:"success_muted"`

	Warning      Color `json:"warning"`
	WarningOn    Color `json:"warning_on"`
	WarningMuted Color `json:"warning_muted"`

	Error      Color `json:"error"`
	ErrorOn    Color `json:"error_on"`
	ErrorMuted Color `json:"error_muted"`

	Info      Color `json:"info"`
	InfoOn    Color `json:"info_on"`
	InfoMuted Color `json:"info_muted"`

	// Miscellaneous UI colors
	Border    Color `json:"border"`
	Outline   Color `json:"outline"`
	Selection Color `json:"selection"`
	Focus     Color `json:"focus"`
	Shadow    Color `json:"shadow"`
	Link      Color `json:"link"`

	Palette Palette `json:"palette"`
}

type Palette struct {
	Red    Color `json:"red"`
	Green  Color `json:"green"`
	Blue   Color `json:"blue"`
	Yellow Color `json:"yellow"`
	Orange Color `json:"orange"`
	Purple Color `json:"purple"`
	Cyan   Color `json:"cyan"`
	Pink   Color `json:"pink"`
	Gray   Color `json:"gray"`
	Brown  Color `json:"brown"`
}

func loadTheme(themeName string, themeDir string) (Theme, error) {
	var theme Theme

	themeFile, err := os.Open(themeDir + "/" + themeName + ".json")
	if err != nil {
		slog.Error("error opening theme file: " + err.Error())
		return theme, err
	}
	ReadJson(themeFile, &theme)

	return theme, nil
}

type Templates struct {
	WallpaperCmd Cmd        `json:"wallpaper_cmd"`
	Templates    []Template `json:"templates"`
}

type Template struct {
	SourceFile Path `json:"source_file"`
	OutputFile Path `json:"output_file"`
	PreHook    Cmd  `json:"pre_hook"`
	PostHook   Cmd  `json:"post_hook"`
}

func initTemplates() (Templates, error) {
	var templates Templates

	ReadJson(*templateFile, &templates)

	return templates, nil
}

func main() {
	kingpin.Parse()

	if *debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	config := loadConfig(*configFile, configFileSetByUser)
	if !themesDirSetByUser && config.ThemesDir != "" {
		*themesDir = string(config.ThemesDir)
	}

	theme, err := loadTheme(*themeName, *themesDir)
	if err != nil {
		return
	}

	templates, err := initTemplates()
	if err != nil {
		slog.Error("error reading templates: " + err.Error())
		return
	}

	out, err := templates.WallpaperCmd.Replace(theme).Run()
	if err != nil {
		slog.Error("error executing WallpaperCmd: " + err.Error())
		return
	}
	slog.Debug("WallpaperCmd out: " + string(out))

	for _, t := range templates.Templates {
		out, err := t.PreHook.Replace(theme).Run()
		if err != nil {
			slog.Error("error executing PreHook: " + err.Error())
			return
		}
		slog.Debug("PreHook out: " + string(out))

		templateFile := t.SourceFile.Resolve(*templatesDir).Open()
		defer templateFile.Close()
		byteTemplate := ReadTemplate(templateFile)
		byteTemplate.Replace(theme)

		outputFile := string(t.OutputFile.Resolve(""))
		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			log.Fatalf("error creating output directory for %s: %s", outputFile, err)
		}
		if err := os.WriteFile(outputFile, byteTemplate, 0644); err != nil {
			log.Fatalf("error writing %s: %s", outputFile, err)
		}

		out, err = t.PostHook.Replace(theme).Run()
		if err != nil {
			slog.Error("error executing PostHook: " + err.Error())
			return
		}
		slog.Debug("PostHook out: " + string(out))

	}
}
