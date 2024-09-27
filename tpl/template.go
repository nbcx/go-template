package tpl

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	logs "log"
)

var (
	beegoTplFuncMap           = make(template.FuncMap)
	beeViewPathTemplateLocked = false
	// beeViewPathTemplates caching map and supported template file extensions per view
	beeViewPathTemplates = make(map[string]map[string]*template.Template)
	templatesLock        sync.RWMutex
	// beeTemplateExt stores the template extension which will build
	beeTemplateExt = []string{"tpl", "html", "gohtml"}
	// beeTemplatePreprocessors stores associations of extension -> preprocessor handler
	beeTemplateEngines = map[string]templatePreProcessor{}
	beeTemplateFS      = defaultFSFunc
)

// ExecuteTemplate applies the template with name  to the specified data object,
// writing the output to wr.
// A template will be executed safely in parallel.
func ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return ExecuteViewPathTemplate(wr, name, Config.ViewsPath, data)
}

// ExecuteViewPathTemplate applies the template with name and from specific viewPath to the specified data object,
// writing the output to wr.
// A template will be executed safely in parallel.
func ExecuteViewPathTemplate(wr io.Writer, name string, viewPath string, data interface{}) error {
	// if BConfig.RunMode == DEV {
	// 	templatesLock.RLock()
	// 	defer templatesLock.RUnlock()
	// }
	if beeTemplates, ok := beeViewPathTemplates[viewPath]; ok {
		if t, ok := beeTemplates[name]; ok {
			var err error
			if t.Lookup(name) != nil {
				err = t.ExecuteTemplate(wr, name, data)
			} else {
				err = t.Execute(wr, data)
			}
			if err != nil {
				logs.Printf("template Execute err: %+v", err)
			}
			return err
		}
		panic("can't find templatefile in the path:" + viewPath + "/" + name)
	}
	panic("Unknown view path:" + viewPath)
}

func init() {
	beegoTplFuncMap["dateformat"] = DateFormat
	beegoTplFuncMap["date"] = Date
	beegoTplFuncMap["compare"] = Compare
	beegoTplFuncMap["compare_not"] = CompareNot
	beegoTplFuncMap["not_nil"] = NotNil
	beegoTplFuncMap["not_null"] = NotNil
	beegoTplFuncMap["substr"] = Substr
	beegoTplFuncMap["html2str"] = HTML2str
	beegoTplFuncMap["str2html"] = Str2html
	beegoTplFuncMap["htmlquote"] = Htmlquote
	beegoTplFuncMap["htmlunquote"] = Htmlunquote
	beegoTplFuncMap["renderform"] = RenderForm
	beegoTplFuncMap["assets_js"] = AssetsJs
	beegoTplFuncMap["assets_css"] = AssetsCSS
	beegoTplFuncMap["config"] = GetConfig
	beegoTplFuncMap["map_get"] = MapGet

	// Comparisons
	beegoTplFuncMap["eq"] = eq // ==
	beegoTplFuncMap["ge"] = ge // >=
	beegoTplFuncMap["gt"] = gt // >
	beegoTplFuncMap["le"] = le // <=
	beegoTplFuncMap["lt"] = lt // <
	beegoTplFuncMap["ne"] = ne // !=
}

// AddFuncMap let user to register a func in the template.
func AddFuncMap(key string, fn interface{}) error {
	beegoTplFuncMap[key] = fn
	return nil
}

type templatePreProcessor func(root, path string, funcs template.FuncMap) (*template.Template, error)

type templateFile struct {
	root  string
	files map[string][]string
}

// visit will make the paths into two part,the first is subDir (without tf.root),the second is full path(without tf.root).
// if tf.root="views" and
// paths is "views/errors/404.html",the subDir will be "errors",the file will be "errors/404.html"
// paths is "views/admin/errors/404.html",the subDir will be "admin/errors",the file will be "admin/errors/404.html"
func (tf *templateFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !HasTemplateExt(paths) {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	file := strings.TrimLeft(replace.Replace(paths[len(tf.root):]), "/")
	subDir := filepath.Dir(file)

	tf.files[subDir] = append(tf.files[subDir], file)
	return nil
}

// HasTemplateExt return this path contains supported template extension of beego or not.
func HasTemplateExt(paths string) bool {
	for _, v := range beeTemplateExt {
		if strings.HasSuffix(paths, "."+v) {
			return true
		}
	}
	return false
}

// AddTemplateExt add new extension for template.
func AddTemplateExt(ext string) {
	for _, v := range beeTemplateExt {
		if v == ext {
			return
		}
	}
	beeTemplateExt = append(beeTemplateExt, ext)
}

// AddViewPath adds a new path to the supported view paths.
// Can later be used by setting a controller ViewPath to this folder
// will panic if called after beego.Run()
func AddViewPath(viewPath string) error {
	if beeViewPathTemplateLocked {
		if _, exist := beeViewPathTemplates[viewPath]; exist {
			return nil // Ignore if viewpath already exists
		}
		panic("Can not add new view paths after beego.Run()")
	}
	beeViewPathTemplates[viewPath] = make(map[string]*template.Template)
	return BuildTemplate(viewPath)
}

func lockViewPaths() {
	beeViewPathTemplateLocked = true
}

func inSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// BuildTemplate will build all template files in a directory.
// it makes beego can render any template file in view directory.
func BuildTemplate(dir string, files ...string) error {
	var err error
	fs := beeTemplateFS()
	f, err := fs.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	defer f.Close()

	beeTemplates, ok := beeViewPathTemplates[dir]
	if !ok {
		panic("Unknown view path: " + dir)
	}
	self := &templateFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err = Walk(fs, dir, self.visit)
	if err != nil {
		fmt.Printf("Walk() returned %v\n", err)
		return err
	}
	buildAllFiles := len(files) == 0
	for _, v := range self.files {
		for _, file := range v {
			if buildAllFiles || inSlice(file, files) {
				templatesLock.Lock()
				ext := filepath.Ext(file)
				var t *template.Template
				if len(ext) == 0 {
					t, err = getTemplate(self.root, fs, file, v...)
				} else if fn, ok := beeTemplateEngines[ext[1:]]; ok {
					t, err = fn(self.root, file, beegoTplFuncMap)
				} else {
					t, err = getTemplate(self.root, fs, file, v...)
				}
				if err != nil {
					logs.Printf("parse template err: %v, %v", file, err)
					templatesLock.Unlock()
					return err
				}
				beeTemplates[file] = t
				templatesLock.Unlock()

				// fmt.Printf("beeTemplates[file]: %+v\n", beeTemplates[file])
			}
		}
	}
	return nil
}

func getTplDeep(root string, fs http.FileSystem, file string, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	var rParent string
	var err error
	if strings.HasPrefix(file, "../") {
		rParent = filepath.Join(filepath.Dir(parent), file)
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		rParent = file
		fileAbsPath = filepath.Join(root, file)
	}
	f, err := fs.Open(fileAbsPath)
	if err != nil {
		panic("can't find template file:" + file)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile(Config.TemplateLeft + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			tl := t.Lookup(m[1])
			if tl != nil {
				continue
			}
			if !HasTemplateExt(m[1]) {
				continue
			}
			_, _, err = getTplDeep(root, fs, m[1], rParent, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func getTemplate(root string, fs http.FileSystem, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims(Config.TemplateLeft, Config.TemplateRight).Funcs(beegoTplFuncMap)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, fs, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = _getTemplate(t, root, fs, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func _getTemplate(t0 *template.Template, root string, fs http.FileSystem, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			// first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = getTplDeep(root, fs, otherFile, "", t)
					if err != nil {
						logs.Panicf("template parse file err:", err)
					} else if len(subMods1) > 0 {
						t, err = _getTemplate(t, root, fs, subMods1, others...)
					}
					break
				}
			}
			// second check define
			for _, otherFile := range others {
				var data []byte
				fileAbsPath := filepath.Join(root, otherFile)
				f, err := fs.Open(fileAbsPath)
				if err != nil {
					f.Close()
					logs.Panicf("template file parse error, not success open file:", err)
					continue
				}
				data, err = io.ReadAll(f)
				f.Close()
				if err != nil {
					logs.Panicf("template file parse error, not success read file:", err)
					continue
				}
				reg := regexp.MustCompile(Config.TemplateLeft + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = getTplDeep(root, fs, otherFile, "", t)
						if err != nil {
							logs.Panicf("template parse file err:", err)
						} else if len(subMods1) > 0 {
							t, err = _getTemplate(t, root, fs, subMods1, others...)
							if err != nil {
								logs.Panicf("template parse file err:", err)
							}
						}
						break
					}
				}
			}
		}
	}
	return
}

type templateFSFunc func() http.FileSystem

func defaultFSFunc() http.FileSystem {
	return FileSystem{}
}

// SetTemplateFSFunc set default filesystem function
func SetTemplateFSFunc(fnt templateFSFunc) {
	beeTemplateFS = fnt
}

// SetViewsPath sets view directory path in beego application.
func SetViewsPath(path string) { // *HttpServer
	Config.ViewsPath = path
	// return BeeApp
}

// SetStaticPath sets static directory path and proper url pattern in beego application.
// if beego.SetStaticPath("static","public"), visit /static/* to load static file in folder "public".
// func SetStaticPath(url string, path string) *HttpServer {
// 	if !strings.HasPrefix(url, "/") {
// 		url = "/" + url
// 	}
// 	if url != "/" {
// 		url = strings.TrimRight(url, "/")
// 	}
// 	BConfig.WebConfig.StaticDir[url] = path
// 	return BeeApp
// }

// DelStaticPath removes the static folder setting in this url pattern in beego application.
// func DelStaticPath(url string) *HttpServer {
// 	if !strings.HasPrefix(url, "/") {
// 		url = "/" + url
// 	}
// 	if url != "/" {
// 		url = strings.TrimRight(url, "/")
// 	}
// 	delete(BConfig.WebConfig.StaticDir, url)
// 	return BeeApp
// }

// AddTemplateEngine add a new templatePreProcessor which support extension
// func AddTemplateEngine(extension string, fn templatePreProcessor) *HttpServer {
// 	AddTemplateExt(extension)
// 	beeTemplateEngines[extension] = fn
// 	return BeeApp
// }
