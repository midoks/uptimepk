package conf

// "net/url"
// "os"

// CustomConf returns the absolute path of custom configuration file that is used.
var CustomConf string

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

// YAMLConfig represents the entire YAML configuration structure
type AppConfig struct {
	AppName   string         `yaml:"app_name"`
	BrandName string         `yaml:"brand_name"`
	RunUser   string         `yaml:"run_user"`
	RunMode   string         `yaml:"run_mode"`
	General   GeneralConfig  `yaml:"general"`
	Admin     Admin          `yaml:"admin"`
	Log       LogConfig      `yaml:"log"`
	Session   SessionConfig  `yaml:"session"`
	Web       WebConfig      `yaml:"web"`
	Security  SecurityConfig `yaml:"security"`
	Database  DatabaseConfig `yaml:"database"`
}

// YAMLConfigSave represents the YAML configuration structure for saving (excludes general and admin)
type YAMLConfigCustom struct {
	AppName   string         `yaml:"app_name"`
	BrandName string         `yaml:"brand_name"`
	RunUser   string         `yaml:"run_user"`
	RunMode   string         `yaml:"run_mode"`
	Log       LogConfig      `yaml:"log"`
	Session   SessionConfig  `yaml:"session"`
	Web       WebConfig      `yaml:"web"`
	Security  SecurityConfig `yaml:"security"`
	Database  DatabaseConfig `yaml:"database"`
}

// GeneralConfig represents the general section in YAML
type GeneralConfig struct {
	MenuFile string `yaml:"menu_file"`
}

// LogConfig represents the log section in YAML
type LogConfig struct {
	Format   string `yaml:"format"`
	RootPath string `yaml:"root_path"`
}

// SessionConfig represents the session section in YAML
type SessionConfig struct {
	Provider       string `yaml:"provider"`
	ProviderConfig string `yaml:"provider_config"`
	CookieName     string `yaml:"cookie_name"`
	CookieSecure   bool   `yaml:"cookie_secure"`
	GCInterval     int64  `yaml:"gc_interval"`
	MaxLifeTime    int64  `yaml:"max_life_time"`
	CSRFCookieName string `yaml:"csrf_cookie_name"`
}

// WebConfig represents the web section in YAML
type WebConfig struct {
	HTTPAddr   string `yaml:"http_addr"`
	HTTPPort   int    `yaml:"http_port"`
	AdminPath  string `yaml:"admin_path"`
	EnableGzip bool   `yaml:"enable_gzip"`
}

// SecurityConfig represents the security section in YAML
type SecurityConfig struct {
	InstallLock             bool   `yaml:"install_lock"`
	SecretKey               string `yaml:"secret_key"`
	LoginRememberDays       int    `yaml:"login_remember_days"`
	CookieRememberName      string `yaml:"cookie_remember_name"`
	CookieUsername          string `yaml:"cookie_username"`
	CookieSecure            bool   `yaml:"cookie_secure"`
	EnableLoginStatusCookie bool   `yaml:"enable_login_status_cookie"`
	LoginStatusCookieName   string `yaml:"login_status_cookie_name"`
}

// DatabaseConfig represents the database section in YAML
type DatabaseConfig struct {
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	DSN         string `yaml:"dsn"`
	TablePrefix string `yaml:"table_prefix"`
	Hostname    string `yaml:"hostname"`
	Hostport    int64  `yaml:"hostport"`
	Name        string `yaml:"name"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	SSLMode     string `yaml:"ssl_mode"`
}

// Admin represents the admin section in YAML
type Admin struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

var (
	App struct {
		// ⚠️ WARNING: Should only be set by the main package (i.e. "imail.go").
		Version string `ini:"-"`

		Name      string
		BrandName string
		RunUser   string
		RunMode   string
		Debug     bool
	}

	// log
	General struct {
		MenuFile string
	}

	// log
	Log struct {
		Format   string
		RootPath string
	}

	// Cache settings
	Cache struct {
		Adapter  string
		Interval int
		Host     string
	}

	// database
	Database struct {
		Type        string `json:"type" env:"TYPE"`
		Path        string `json:"path" env:"PATH"`
		DSN         string `json:"dsn" env:"DSN"`
		TablePrefix string `json:"table_prefix" env:"TABLE_PREFIX"`
		Hostname    string `json:"hostname" env:"HOST"`
		Hostport    int64  `json:"hostport" env:"PORT"`
		Name        string `json:"name" env:"NAME"`
		User        string `json:"user" env:"USER"`
		Password    string `json:"password" env:"PASS"`
		SSLMode     string `json:"ssl_mode" env:"SSL_MODE"`
	}

	// web settings
	Web struct {
		HTTPAddr   string
		HTTPPort   int
		AdminPath  string
		EnableGzip bool
	}

	Session struct {
		Provider       string
		ProviderConfig string
		CookieName     string
		CookieSecure   bool
		GCInterval     int64
		MaxLifeTime    int64
		CSRFCookieName string
	}

	// Security settings
	Security struct {
		InstallLock             bool
		SecretKey               string
		LoginRememberDays       int
		CookieRememberName      string
		CookieUsername          string
		CookieSecure            bool
		EnableLoginStatusCookie bool
		LoginStatusCookieName   string
	}
)
