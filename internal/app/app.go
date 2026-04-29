package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	// "time"
	// gbrotli "github.com/anargu/gin-brotli"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"uptimepk/embed"
	"uptimepk/internal/app/middleware"
	"uptimepk/internal/conf"
	"uptimepk/internal/op"

	backend "uptimepk/internal/app/handles/backend"
	backend_admin "uptimepk/internal/app/handles/backend/admin"
	backend_log "uptimepk/internal/app/handles/backend/log"
	backend_monitor "uptimepk/internal/app/handles/backend/monitor"
	backend_system "uptimepk/internal/app/handles/backend/system"
	"uptimepk/internal/app/handles/home"
	"uptimepk/internal/app/handles/install"
)

func initTmplFunc(r *gin.Engine) {
	// Define template functions
	funcMap := template.FuncMap{
		"safe": func(str string) template.HTML {
			return template.HTML(str)
		},
		// Cache-busting token exposed as a function for templates
		"BuildCommit": func() string {
			return conf.BuildCommit
		},
		"HasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		//是子菜单或当前菜单
		"IsSubOrEq": func(base, menu string) bool {
			if base == menu {
				return true
			}
			endp := strings.Replace(base, menu, "", 1)
			endp = strings.TrimPrefix(endp, "/")
			return !strings.Contains(endp, "/")
		},
		"Contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"json": func(v interface{}) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(b)
		},
		"formatBytes": func(bytes int64) string {
			if bytes < 1024 {
				return fmt.Sprintf("%d B", bytes)
			}
			return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
		},
	}

	// Build template set with directory-aware names (e.g., "install/index.tmpl")
	// so that we can reference templates across multiple directories explicitly.
	tpl := template.New("").Delims("{[", "]}").Funcs(funcMap)

	for _, name := range embed.TemplatesAllNames("templates") {
		// Trim the leading "templates/" so template names are like "install/index.tmpl"
		short := strings.TrimPrefix(name, "templates/")
		content, err := embed.Templates.ReadFile(name)
		if err != nil {
			panic(err)
		}
		if _, err := tpl.New(short).Parse(string(content)); err != nil {
			panic(err)
		}
	}

	r.SetHTMLTemplate(tpl)
}

// 后台/backstage
func initRuoteAdmin(r *gin.Engine) {
	// fmt.Println("conf.Web.AdminPath:", conf.Web.AdminPath)
	backstage := r.Group(conf.Web.AdminPath)
	backstage.Use(middleware.CheckInstalled())
	backstage.GET("/login", backend.LoginPage)
	backstage.POST("/login", backend.PostLogin)
	backstage.GET("/logout", backend.LoginOut)

	backstage_admin := backstage.Group("")
	backstage_admin.Use(middleware.CheckInstalled(), middleware.AuthRequired())

	// 管理员
	// backstage_admin.GET("", backend.HomePage)
	backstage_admin.GET("", backend_monitor.Home)
	backstage_admin.GET("/index", backend_monitor.Home)

	backstage_admin.GET("/admin/index", backend_admin.Home)
	backstage_admin.GET("/admin/add", backend_admin.Add)
	backstage_admin.POST("/admin/add", backend_admin.PostAdd)
	backstage_admin.GET("/admin/list", backend_admin.List)
	backstage_admin.GET("/admin/details", backend_admin.Details)
	backstage_admin.GET("/admin/update", backend_admin.Update)
	backstage_admin.POST("/admin/delete", backend_admin.Delete)
	backstage_admin.POST("/admin/trigger_status", backend_admin.AdminTriggerStatus)

	// 管理员 - 通知
	backstage_admin.GET("/admin/recipients", backend_admin.Recipients)
	backstage_admin.GET("/admin/recipients/list", backend_admin.RecipientsList)
	backstage_admin.POST("/admin/recipients/delete", backend_admin.RecipientsDelete)
	backstage_admin.GET("/admin/recipients/add", backend_admin.RecipientsAdd)
	backstage_admin.POST("/admin/recipients/add", backend_admin.PostRecipientsAdd)
	backstage_admin.GET("/admin/recipients/groups", backend_admin.RecipientsGroups)
	backstage_admin.GET("/admin/recipients/groups/list", backend_admin.RecipientsGroupsList)
	backstage_admin.GET("/admin/recipients/groups/select", backend_admin.RecipientsGroupsSelect)
	backstage_admin.GET("/admin/recipients/groups/add", backend_admin.RecipientsGroupsAdd)
	backstage_admin.POST("/admin/recipients/groups/add", backend_admin.PostRecipientsGroupsAdd)
	backstage_admin.POST("/admin/recipients/groups/delete", backend_admin.PostRecipientsGroupsDelete)
	backstage_admin.GET("/admin/recipients/instances", backend_admin.RecipientsInstances)
	backstage_admin.GET("/admin/recipients/instances/list", backend_admin.RecipientsInstancesList)
	backstage_admin.GET("/admin/recipients/instances/add", backend_admin.RecipientsInstancesAdd)
	backstage_admin.POST("/admin/recipients/instances/add", backend_admin.PostRecipientsInstancesAdd)
	backstage_admin.GET("/admin/recipients/instances/details", backend_admin.RecipientsInstancesDetails)
	backstage_admin.GET("/admin/recipients/instances/update", backend_admin.RecipientsInstancesUpdate)
	backstage_admin.GET("/admin/recipients/instances/test", backend_admin.RecipientsInstancesTest)
	backstage_admin.POST("/admin/recipients/instances/test", backend_admin.PostRecipientsInstancesTest)
	backstage_admin.POST("/admin/recipients/instances/delete", backend_admin.RecipientsInstancesDelete)

	backstage_admin.GET("/admin/recipients/recipients/details", backend_admin.RecipientsRecipientsDetails)
	backstage_admin.GET("/admin/recipients/recipients/update", backend_admin.RecipientsRecipientsUpdate)
	backstage_admin.GET("/admin/recipients/recipients/test", backend_admin.RecipientsRecipientsTest)

	backstage_admin.GET("/admin/recipients/tasks", backend_admin.RecipientsTasks)
	backstage_admin.GET("/admin/recipients/logs", backend_admin.RecipientsLogs)

	// 监控管理
	backstage_admin.GET("/monitor", backend_monitor.Home)
	backstage_admin.GET("/monitor/add", backend_monitor.Add)
	backstage_admin.POST("/monitor/add", backend_monitor.PostAdd)
	backstage_admin.GET("/monitor/list", backend_monitor.List)
	backstage_admin.POST("/monitor/delete", backend_monitor.SoftDelete)
	backstage_admin.POST("/monitor/trigger_status", backend_monitor.MonitorTriggerStatus)

	backstage_admin.GET("/monitor/group", backend_monitor.MonitorGroups)
	backstage_admin.GET("/monitor/group/add", backend_monitor.MonitorGroupsAdd)
	backstage_admin.GET("/monitor/group/list", backend_monitor.MonitorGroupsList)
	backstage_admin.POST("/monitor/group/add", backend_monitor.PostMonitorGroupsAdd)
	backstage_admin.POST("/monitor/group/delete", backend_monitor.MonitorGroupsDelete)
	backstage_admin.POST("/monitor/group/trigger_status", backend_monitor.MonitorGroupsTriggerStatus)
	backstage_admin.POST("/monitor/group/sort", backend_monitor.MonitorGroupsSort)

	backstage_admin.GET("/monitor/log", backend_monitor.MonitorLog)
	backstage_admin.GET("/monitor/log/list", backend_monitor.MonitorLogList)
	backstage_admin.POST("/monitor/log/delete", backend_monitor.MonitorLogDelete)

	// 日志审计
	backstage_admin.GET("/log", backend_log.Home)
	backstage_admin.GET("/log/list", backend_log.List)
	backstage_admin.GET("/log/settings", backend_log.Settings)
	backstage_admin.POST("/log/settings", backend_log.PostSettting)
	backstage_admin.GET("/log/clean", backend_log.Clean)
	backstage_admin.POST("/log/clean", backend_log.PostLogClean)
	backstage_admin.POST("/log/delete", backend_log.Delete)

	// 系统设置
	backstage_admin.GET("/system/settings", backend_system.Home)
	backstage_admin.POST("/system/settings/home", backend_system.PostHome)

	backstage_admin.GET("/system/settings/web", backend_system.Web)
	backstage_admin.POST("/system/settings/web", backend_system.PostWeb)

	backstage_admin.GET("/system/settings/profile", backend_system.Profile)
	backstage_admin.POST("/system/settings/profile", backend_system.PostProfile)

	backstage_admin.GET("/system/settings/login", backend_system.Login)
	backstage_admin.POST("/system/settings/login", backend_system.PostLogin)
	backstage_admin.GET("/system/settings/login/logs", backend_system.LoginLogs)
	backstage_admin.GET("/system/settings/login/logs/list", backend_system.LoginLogsList)

	backstage_admin.GET("/system/database", backend_system.Database)
	backstage_admin.GET("/system/database/index", backend_system.Database)
	backstage_admin.GET("/system/database/list", backend_system.DatabaseList)
	backstage_admin.GET("/system/database/update", backend_system.DatabaseUpdate)
	backstage_admin.GET("/system/database/cleans", backend_system.DatabaseClean)
	backstage_admin.POST("/system/database/clean", backend_system.PostDatabaseClean)
	backstage_admin.POST("/system/database/delete", backend_system.PostDatabaseDelete)

	backstage_admin.GET("/system/database/clean_setting", backend_system.DatabaseCleanSetting)
	backstage_admin.POST("/system/database/clean_setting", backend_system.PostDatabaseCleanSetting)
	backstage_admin.GET("/system/db", backend_system.Db)
	backstage_admin.GET("/system/db/list", backend_system.DbNodeList)
	backstage_admin.GET("/system/db/add", backend_system.DbNodeAdd)
	backstage_admin.POST("/system/db/add", backend_system.PostDbNodeAdd)
	backstage_admin.GET("/system/db/details", backend_system.DbNodeDetails)
	backstage_admin.GET("/system/db/clean", backend_system.DbNodeClean)
	backstage_admin.GET("/system/db/logs", backend_system.DbNodeLogs)
	backstage_admin.GET("/system/db/update", backend_system.DbNodeUpdate)

}

func initRuoteInstall(r *gin.Engine) {
	installGroup := r.Group("/install")
	installGroup.Use(middleware.CheckInstalledAfter())
	installGroup.GET("/index", install.HomePage)
	installGroup.POST("/step1", install.PostInstallStep1)
	installGroup.POST("/dbtest", install.MyDbtest)
}

func initRuoteFrontend(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.Use(middleware.CheckInstalled()).GET("/", home.Index)
	r.Use(middleware.CheckInstalled()).GET("/groups", home.Groups)
	r.Use(middleware.CheckInstalled()).GET("/monitor", home.Monitor)
	r.GET("/ws/status", home.WSHandler)
}

func initRuote(r *gin.Engine) {
	// static files from embedded filesystem subdir "static"
	staticFS, err := fs.Sub(embed.Static, "static")
	if err != nil {
		panic(fmt.Sprintf("initRuote:%v", err))
	}
	// 设置静态文件缓存为一周
	staticHandler := http.StripPrefix("/static", http.FileServer(http.FS(staticFS)))
	r.GET("/static/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=604800") // 604800 秒 = 7 天
		// c.Header("Permissions-Policy", "unload=*")
		staticHandler.ServeHTTP(c.Writer, c.Request)
	})

	initRuoteInstall(r)
	initRuoteAdmin(r)
	initRuoteFrontend(r)
}

func Run() {
	if conf.App.RunMode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化清理任务
	op.InitCleanTask()

	// 初始化接收人汇总任务
	op.InitRecipientsSummaryTasks()

	r := gin.New()

	// 初始化 session 存储
	store := cookie.NewStore([]byte(conf.Security.SecretKey))
	// 设置 cookie 选项
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(conf.Session.MaxLifeTime),
		HttpOnly: true,
		Secure:   conf.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(conf.Session.CookieName, store))

	// 启用压缩
	if conf.Web.EnableGzip {
		// 先添加 brotli 压缩（比 gzip 压缩率更高）
		// r.Use(gbrotli.Brotli(gbrotli.DefaultCompression))
		// 再添加 gzip 压缩作为 fallback
		r.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	// 设置 Permissions-Policy 头，允许 unload 事件
	// r.Use(func(c *gin.Context) {
	// 	c.Header("Permissions-Policy", "unload=*")
	// 	c.Next()
	// })

	// if conf.App.Debug {
	// 	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 		p := param.Path
	// 		if strings.Contains(p, ".js") || strings.Contains(p, ".css") {
	// 			return ""
	// 		}
	// 		if strings.Contains(p, ".woff2") {
	// 			return ""
	// 		}
	// 		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s \"%s\"\n",
	// 			param.ClientIP,
	// 			param.TimeStamp.Format(time.RFC1123),
	// 			param.Method,
	// 			p,
	// 			param.Request.Proto,
	// 			param.StatusCode,
	// 			param.Latency,
	// 			param.ErrorMessage,
	// 		)
	// 	}))
	// }

	// r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	initTmplFunc(r)
	initRuote(r)
	r.Run(fmt.Sprintf(":%d", conf.Web.HTTPPort))
}
