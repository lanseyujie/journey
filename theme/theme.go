package theme

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

type Theme struct {
    name      string                        // theme name
    root      string                        // template storage root path
    extension string                        // template file extension
    debug     bool                          // disable cache flag
    files     []string                      // template files
    cache     map[string]*template.Template // TplName => TplObj
    sync.RWMutex
}

// NewTheme
func NewTheme(name, root string) *Theme {
    return &Theme{
        name:      name,
        root:      root,
        extension: "html",
        cache:     make(map[string]*template.Template),
    }
}

// Extension
func (t *Theme) Extension(ext string) *Theme {
    t.extension = ext

    return t
}

// DisableCache
func (t *Theme) DisableCache(flag ...bool) *Theme {
    if len(flag) > 0 {
        t.debug = flag[0]
    } else {
        t.debug = true
    }

    return t
}

// Index supported template files
func (t *Theme) Index() error {
    t.files = []string{}

    err := filepath.Walk(t.root, func(path string, f os.FileInfo, err error) error {
        if f == nil {
            return err
        }

        // ignore directories and symbolic links
        if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
            return nil
        }

        if !strings.HasSuffix(path, "."+t.extension) {
            return nil
        }

        // save the relative path of the template file
        t.files = append(t.files, strings.TrimLeft(path[len(t.root):], "/"))

        return nil
    })

    return err
}

// Parse template
func (t *Theme) Parse(file, parent string, tpl *template.Template) (*template.Template, error) {
    var path string
    if strings.HasPrefix(file, "../") {
        parent = filepath.Join(filepath.Dir(parent), file)
        path = filepath.Join(t.root, filepath.Dir(parent), file)
    } else {
        parent = file
        path = filepath.Join(t.root, file)
    }

    html, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    if tpl == nil {
        tpl, err = template.New(file).Funcs(funcMap).Parse(string(html))
    } else {
        tpl, err = tpl.New(file).Parse(string(html))
    }

    if err != nil {
        return nil, err
    }

    // get the file name of the sub-template
    re := regexp.MustCompile(`{{[ ]*template[ ]+"([^"]+)"`)
    // e.g. [ [ {{template "../header.html" ../header.html ] ]
    allSubTpl := re.FindAllStringSubmatch(string(html), -1)
    for _, sub := range allSubTpl {
        if len(sub) == 2 {
            // ignore the template associated with tpl
            if tpl.Lookup(sub[1]) != nil {
                continue
            }

            // ignore the file name without extension, e.g. {{template "banner" .}}
            if !strings.HasSuffix(sub[1], "."+t.extension) {
                continue
            }

            // parse sub-template
            _, err = t.Parse(sub[1], parent, tpl)
            if err != nil {
                return nil, err
            }
        }
    }

    return tpl, nil
}

// Build the template with the relative path of the file
func (t *Theme) Build(files ...string) (err error) {
    _, err = os.Stat(t.root)
    if err != nil {
        return
    }

    // build all templates in the theme directory if no file is specified
    if len(files) == 0 {
        err = t.Index()
        if err != nil {
            return
        }
        files = t.files
    }

    t.Lock()
    for _, file := range files {
        var tpl *template.Template
        tpl, err = t.Parse(file, "", nil)
        if err != nil {
            err = errors.New("template: " + file + " parse error, " + err.Error())
            break
        }
        t.cache[file] = tpl
    }
    t.Unlock()

    return
}

// Render
func (t *Theme) Render(wr io.Writer, name string, data interface{}) (err error) {
    if t.debug {
        err = t.Build(name)
        if err != nil {
            return
        }
    }

    t.RLock()
    if tpl := t.cache[name]; tpl != nil {
        if tpl.Lookup(name) != nil {
            err = tpl.ExecuteTemplate(wr, name, data)
        } else {
            err = tpl.Execute(wr, data)
        }
    } else {
        err = errors.New("template: " + t.name + " not found")
    }
    t.RUnlock()

    return
}

// Render template with layout file
func (t *Theme) RenderLayout(wr io.Writer, name, layout string, data map[string]interface{}) error {
    var buf bytes.Buffer
    err := t.Render(&buf, name, data)
    if err != nil {
        return err
    }

    data["LayoutContent"] = template.HTML(buf.String())

    // layout
    return t.Render(wr, layout, data)
}
