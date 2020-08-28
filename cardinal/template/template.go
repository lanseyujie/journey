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

type Template struct {
    name      string                        // theme name
    root      string                        // template storage root path
    extension string                        // template file extension
    debug     bool                          // disable cache flag
    files     []string                      // template files
    cache     map[string]*template.Template // TplName => TplObj
    sync.RWMutex
}

// NewTemplate
func NewTemplate(name, root string) *Template {
    return &Template{
        name:      name,
        root:      root,
        extension: "html",
        cache:     make(map[string]*template.Template),
    }
}

// Extension
func (tpl *Template) Extension(ext string) *Template {
    tpl.extension = ext

    return tpl
}

// DisableCache
func (tpl *Template) DisableCache(flag ...bool) *Template {
    if len(flag) > 0 {
        tpl.debug = flag[0]
    } else {
        tpl.debug = true
    }

    return tpl
}

// Index supported template files
func (tpl *Template) Index() error {
    tpl.files = []string{}

    err := filepath.Walk(tpl.root, func(path string, f os.FileInfo, err error) error {
        if f == nil {
            return err
        }

        // ignore directories and symbolic links
        if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
            return nil
        }

        if !strings.HasSuffix(path, "."+tpl.extension) {
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
    var path string
    if strings.HasPrefix(file, "../") {
        parent = filepath.Join(filepath.Dir(parent), file)
        path = filepath.Join(tpl.root, filepath.Dir(parent), file)
    } else {
        parent = file
        path = filepath.Join(tpl.root, file)
    }

    html, err := ioutil.ReadFile(path)
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
    re := regexp.MustCompile(`{{[ ]*template[ ]+"([^"]+)"`)
    // e.g. [ [ {{template "../header.html" ../header.html ] ]
    allSubTpl := re.FindAllStringSubmatch(string(html), -1)
    for _, sub := range allSubTpl {
        if len(sub) == 2 {
            // ignore the template associated with t
            if t.Lookup(sub[1]) != nil {
                continue
            }

            // ignore the file name without extension, e.g. {{template "banner" .}}
            if !strings.HasSuffix(sub[1], "."+tpl.extension) {
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
func (tpl *Template) Build(files ...string) (err error) {
    _, err = os.Stat(tpl.root)
    if err != nil {
        return
    }

    // build all templates in the theme directory if no file is specified
    if len(files) == 0 {
        err = tpl.Index()
        if err != nil {
            return
        }
        files = tpl.files
    }

    tpl.Lock()
    for _, file := range files {
        var t *template.Template
        t, err = tpl.Parse(file, "", nil)
        if err != nil {
            err = errors.New("template: " + file + " parse error, " + err.Error())
            break
        }
        tpl.cache[file] = t
    }
    tpl.Unlock()

    return
}

// Render
func (tpl *Template) Render(wr io.Writer, name string, data interface{}) (err error) {
    if tpl.debug {
        err = tpl.Build(name)
        if err != nil {
            return
        }
    }

    tpl.RLock()
    if t := tpl.cache[name]; t != nil {
        if t.Lookup(name) != nil {
            err = t.ExecuteTemplate(wr, name, data)
        } else {
            err = t.Execute(wr, data)
        }
    } else {
        err = errors.New("template: " + tpl.name + " not found")
    }
    tpl.RUnlock()

    return
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
