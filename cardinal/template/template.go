package template

import (
    "bytes"
    "errors"
    "html/template"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "sync"
)

type ThemeCache map[string]*template.Template // TplName => TplObj

type Template struct {
    name  string // theme name
    root  string // template storage root path
    files []string
}

var (
    extension    = "html" // template file extension
    disableCache = false
    funcMap      = make(template.FuncMap)
    pool         = make([]*Template, 0, 1)
    cache        = make(map[string]ThemeCache) // ThemeName => ThemeCache
    cacheLock    = &sync.RWMutex{}
)

func init() {
    AddFuncMap("html", Html)
    AddFuncMap("string", String)
    AddFuncMap("stringjoin", StringJoin)
    AddFuncMap("dateformat", DateFormat)
    AddFuncMap("substr", Substr)
    AddFuncMap("add", Add)
    AddFuncMap("sub", Subtract)
    AddFuncMap("mul", Multiply)
    AddFuncMap("div", Divide)
}

// NewTemplate
func NewTemplate(name, root string) *Template {
    tpl := Theme(name)
    if tpl != nil {
        return tpl
    }

    tpl = &Template{
        name:  name,
        root:  root,
        files: make([]string, 0, 4),
    }
    pool = append(pool, tpl)

    return tpl
}

// Extension
func Extension(ext string) {
    extension = ext
}

// DisableCache
func DisableCache(disable bool) {
    disableCache = disable
}

// AddFuncMap register a func in the template
func AddFuncMap(key string, fn interface{}) {
    funcMap[key] = fn
}

// Theme
func Theme(name string) *Template {
    for _, t := range pool {
        if t != nil && t.name == name {
            return t
        }
    }

    return nil
}

// Index supported template files
func (tpl *Template) Index() error {
    if disableCache {
        tpl.files = make([]string, 0, 4)
    }

    err := filepath.Walk(tpl.root, func(path string, f os.FileInfo, err error) error {
        if f == nil {
            return err
        }

        // ignore directories and symbolic links
        if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
            return nil
        }

        if !strings.HasSuffix(path, "."+extension) {
            return nil
        }

        // save the relative path of the template file
        tpl.files = append(tpl.files, strings.TrimLeft(path[len(tpl.root):], "/"))

        return nil
    })

    return err
}

// Parse template
func (tpl *Template) Parse(file, parent string, t *template.Template) (*template.Template, error) {
    var abspath string
    if strings.HasPrefix(file, "../") {
        parent = filepath.Join(filepath.Dir(parent), file)
        abspath = filepath.Join(tpl.root, filepath.Dir(parent), file)
    } else {
        parent = file
        abspath = filepath.Join(tpl.root, file)
    }

    html, err := ioutil.ReadFile(abspath)
    if err != nil {
        return nil, err
    }

    if t == nil {
        t, err = template.New(file).Funcs(funcMap).Parse(string(html))
    } else {
        t, err = t.New(file).Parse(string(html))
    }
    if err != nil {
        return nil, err
    }

    // get the file name of the sub-template
    re := regexp.MustCompile("{{[ ]*template[ ]+\"([^\"]+)\"")
    // e.g. [ [ {{template "../header.html" ../header.html ] ]
    allSubTpl := re.FindAllStringSubmatch(string(html), -1)
    for _, sub := range allSubTpl {
        if len(sub) == 2 {
            // ignore the template associated with t
            if t.Lookup(sub[1]) != nil {
                continue
            }

            // ignore the file name without extension, e.g. {{template "banner" .}}
            if !strings.HasSuffix(sub[1], "."+extension) {
                continue
            }

            // parse sub-template
            _, err = tpl.Parse(sub[1], parent, t)
            if err != nil {
                return nil, err
            }
        }
    }

    return t, nil
}

// Build the template with the relative path of the file
func (tpl *Template) Build(files ...string) error {
    if _, err := os.Stat(tpl.root); err != nil {
        return errors.New("template: " + err.Error())
    }

    theme := make(ThemeCache)

    // build all templates in the theme directory if no file is specified
    if len(files) == 0 {
        if err := tpl.Index(); err != nil {
            return errors.New("template: index error, " + err.Error())
        }
        files = tpl.files
    }

    for _, file := range files {
        t, err := tpl.Parse(file, "", nil)
        if err != nil {
            return errors.New("template: parse error, " + err.Error())
        }
        theme[file] = t
    }

    cacheLock.Lock()
    cache[tpl.name] = theme
    cacheLock.Unlock()

    return nil
}

// Render
func (tpl *Template) Render(wr io.Writer, name string, data interface{}) (err error) {
    if disableCache {
        err = tpl.Build(name)
        if err != nil {
            return
        }
    }

    if c, exist := cache[tpl.name]; exist {
        if t := c[name]; t != nil {
            if t.Lookup(name) != nil {
                err = t.ExecuteTemplate(wr, name, data)
            } else {
                err = t.Execute(wr, data)
            }

            return
        }

        panic("template file not found:" + tpl.name + "/" + name)
    }

    panic("theme not found:" + tpl.name)
}

// Render
func Render(wr io.Writer, theme, name string, data interface{}) (err error) {
    tpl := Theme(theme)
    if tpl != nil {
        return tpl.Render(wr, name, data)
    }

    panic("theme not found:" + theme)
}

// Render template with layout file
func (tpl *Template) RenderLayout(wr io.Writer, name, layout string, data map[string]interface{}) error {
    var buf bytes.Buffer
    err := tpl.Render(&buf, name, data)
    if err != nil {
        return err
    }

    data["LayoutContent"] = template.HTML(buf.String())

    // layout
    return tpl.Render(wr, layout, data)
}

// Render template with layout file
func RenderLayout(wr io.Writer, theme, name, layout string, data map[string]interface{}) error {
    tpl := Theme(theme)
    if tpl != nil {
        return tpl.RenderLayout(wr, name, layout, data)
    }

    panic("theme not found:" + theme)
}
