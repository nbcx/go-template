package tpl

const (
	// DEV is for develop
	DEV = "dev"
	// PROD is for production
	PROD = "prod"
)

// WebConfig holds web related config
type TplConfig struct {
	// RunMode
	// @Description it's the same as environment. In general, we have different run modes.
	// For example, the most common case is using dev, test, prod three environments
	// when you are developing the application, you should set it as dev
	// when you completed coding and want QA to test your code, you should deploy your application to test environment
	// and the RunMode should be set as test
	// when you completed all tests, you want to deploy it to prod, you should set it to prod
	// You should never set RunMode="dev" when you deploy the application to prod
	// because Beego will do more things which need Go SDK and other tools when it found out the RunMode="dev"
	// @Default dev
	RunMode string // Running Mode: dev | prod
	// AutoRender
	// @Description If it's true, Beego will render the page based on your template and data
	// In general, keep it as true.
	// But if you are building RESTFul API and you don't have any page,
	// you can set it to false
	// @Default true
	AutoRender bool
	// Deprecated: Beego didn't use it anymore
	EnableDocs bool
	// EnableXSRF
	// @Description If it's true, Beego will help to provide XSRF support
	// But you should notice that, now Beego only work for HTTPS protocol with XSRF
	// because it's not safe if using HTTP protocol
	// And, the cookie storing XSRF token has two more flags HttpOnly and Secure
	// It means that you must use HTTPS protocol and you can not read the token from JS script
	// This is completed different from Beego 1.x because we got many security reports
	// And if you are in dev environment, you could set it to false
	// @Default false
	EnableXSRF bool
	// DirectoryIndex
	// @Description When Beego serves static resources request, it will look up the file.
	// If the file is directory, Beego will try to find the index.html as the response
	// But if the index.html is not exist or it's a directory,
	// Beego will return 403 response if DirectoryIndex is **false**
	// @Default false
	DirectoryIndex bool
	// FlashName
	// @Description the cookie's name when Beego try to store the flash data into cookie
	// @Default BEEGO_FLASH
	FlashName string
	// FlashSeparator
	// @Description When Beego read flash data from request, it uses this as the separator
	// @Default BEEGOFLASH
	FlashSeparator string
	// StaticDir
	// @Description Beego uses this as static resources' root directory.
	// It means that Beego will try to search static resource from this start point
	// It's a map, the key is the path and the value is the directory
	// For example, the default value is /static => static,
	// which means that when Beego got a request with path /static/xxx
	// Beego will try to find the resource from static directory
	// @Default /static => static
	// StaticDir map[string]string
	// StaticExtensionsToGzip
	// @Description The static resources with those extension will be compressed if EnableGzip is true
	// @Default [".css", ".js" ]
	StaticExtensionsToGzip []string
	// StaticCacheFileSize
	// @Description If the size of static resource < StaticCacheFileSize, Beego will try to handle it by itself,
	// it means that Beego will compressed the file data (if enable) and cache this file.
	// But if the file size > StaticCacheFileSize, Beego just simply delegate the request to http.ServeFile
	// the default value is 100KB.
	// the max memory size of caching static files is StaticCacheFileSize * StaticCacheFileNum
	// see StaticCacheFileNum
	// @Default 102400
	StaticCacheFileSize int
	// StaticCacheFileNum
	// @Description Beego use it to control the memory usage of caching static resource file
	// If the caching files > StaticCacheFileNum, Beego use LRU algorithm to remove caching file
	// the max memory size of caching static files is StaticCacheFileSize * StaticCacheFileNum
	// see StaticCacheFileSize
	// @Default 1000
	StaticCacheFileNum int
	// TemplateLeft
	// @Description Beego use this to render page
	// see TemplateRight
	// @Default {{
	TemplateLeft string
	// TemplateRight
	// @Description Beego use this to render page
	// see TemplateLeft
	// @Default }}
	TemplateRight string
	// ViewsPath
	// @Description The directory of Beego application storing template
	// @Default views
	ViewsPath string
	// CommentRouterPath
	// @Description Beego scans this directory and its sub directory to generate router
	// Beego only scans this directory when it's in dev environment
	// @Default controllers
	CommentRouterPath string
	// XSRFKey
	// @Description the name of cookie storing XSRF token
	// see EnableXSRF
	// @Default beegoxsrf
	XSRFKey string
	// XSRFExpire
	// @Description the expiration time of XSRF token cookie
	// second
	// @Default 0
	XSRFExpire int
}

var (
	// BConfig is the default config for Application
	Config *TplConfig
)

func init() {
	Config = newBConfig()
}

func newBConfig() *TplConfig {
	res := &TplConfig{
		// RunMode:             PROD,
		AutoRender:     true,
		EnableDocs:     false,
		FlashName:      "BEEGO_FLASH",
		FlashSeparator: "BEEGOFLASH",
		DirectoryIndex: false,
		// StaticDir:              map[string]string{"/static": "static"},
		StaticExtensionsToGzip: []string{".css", ".js"},
		StaticCacheFileSize:    1024 * 100,
		StaticCacheFileNum:     1000,
		TemplateLeft:           "{{",
		TemplateRight:          "}}",
		ViewsPath:              "views",
		CommentRouterPath:      "controllers",
		EnableXSRF:             false,
		XSRFKey:                "beegoxsrf",
		XSRFExpire:             0,
	}
	return res
}
