package tpl

import (
	"bytes"
	"net/http"
)

// Controller defines some basic http request handler operations, such as
// http context, template and view, session and xsrf.
// type Controller struct {
// 	// context data
// 	Data map[interface{}]interface{}

// 	// route controller info
// 	controllerName string
// 	actionName     string
// 	// methodMapping  map[string]func() //method:routertree
// 	// AppController interface{}

// 	// template data
// 	// TplName        string
// 	ViewPath string
// 	// Layout         string
// 	LayoutSections map[string]string // the key is the section name and the value is the template name
// 	TplPrefix      string
// 	// TplExt         string
// 	EnableRender bool

// 	// xsrf data
// 	// EnableXSRF bool
// 	// _xsrfToken string
// 	XSRFExpire int
// }

// Init generates default values of controller operations.
// func (c *Controller) Init(ctx *hi.Context) {
// 	// c.Layout = ""
// 	// c.TplName = ""
// 	c.controllerName = "ControllerName"
// 	c.actionName = "actionName"
// 	// c.TplExt = "tpl"
// 	c.EnableRender = true
// 	// ctx.EnableXSRF = true
// 	c.Data = map[interface{}]interface{}{}
// }

// Render sends the response with rendered template bytes as text/html type.
func Render(w http.ResponseWriter, tplName string, data map[interface{}]interface{}) error {
	// c.TplName = tpl
	rb, err := RenderBytes(tplName, data)

	if err != nil {
		return err
	}

	// w := ctx.Writer
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	_, err = w.Write(rb)
	return err
}

// RenderString returns the rendered template string. Do not send out response.
func RenderString(tplName string, data map[interface{}]interface{}) (string, error) {
	b, e := RenderBytes(tplName, data)
	if e != nil {
		return "", e
	}
	return string(b), e
}

// RenderBytes returns the bytes of rendered template string. Do not send out response.
func RenderBytes(tplName string, data map[interface{}]interface{}) ([]byte, error) {
	buf, err := renderTemplate(tplName, data)
	// if the controller has set layout, then first get the tplName's content set the content to the layout
	// if err == nil && c.Layout != "" {
	// 	c.Data["LayoutContent"] = template.HTML(buf.String())

	// 	if c.LayoutSections != nil {
	// 		for sectionName, sectionTpl := range c.LayoutSections {
	// 			if sectionTpl == "" {
	// 				c.Data[sectionName] = ""
	// 				continue
	// 			}
	// 			buf.Reset()
	// 			err = ExecuteViewPathTemplate(&buf, sectionTpl, Config.ViewsPath, c.Data)
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			c.Data[sectionName] = template.HTML(buf.String())
	// 		}
	// 	}

	// 	buf.Reset()
	// 	err = ExecuteViewPathTemplate(&buf, c.Layout, Config.ViewsPath, c.Data)
	// }
	return buf.Bytes(), err
}

func renderTemplate(tplName string, data map[interface{}]interface{}) (bytes.Buffer, error) {
	var buf bytes.Buffer
	// if c.TplName == "" {
	// 	c.TplName = strings.ToLower(c.controllerName) + "/" + strings.ToLower(c.actionName) + "." + c.TplExt
	// }
	// if c.TplPrefix != "" {
	// 	tplName = c.TplPrefix + tplName
	// }
	if Config.RunMode == DEV {
		buildFiles := []string{tplName}
		// if c.Layout != "" {
		// 	buildFiles = append(buildFiles, c.Layout)
		// 	if c.LayoutSections != nil {
		// 		for _, sectionTpl := range c.LayoutSections {
		// 			if sectionTpl == "" {
		// 				continue
		// 			}
		// 			buildFiles = append(buildFiles, sectionTpl)
		// 		}
		// 	}
		// }
		_ = BuildTemplate(Config.ViewsPath, buildFiles...)
	}
	return buf, ExecuteViewPathTemplate(&buf, tplName, Config.ViewsPath, data)
}

// func (c *Controller) viewPath() string {
// 	if c.ViewPath == "" {
// 		return Config.ViewsPath
// 	}
// 	return c.ViewPath
// }
