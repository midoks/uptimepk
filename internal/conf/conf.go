package conf

import (

	// "net/url"
	"os"
	"path/filepath"
	"strconv"

	// "strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"uptimepk/embed"
)

var appConfig AppConfig

func ReadConf() error {
	data, err := embed.Conf.ReadFile("conf/app.yaml")
	if err != nil {
		return errors.Wrap(err, "read file 'conf/app.yaml'")
	}

	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		return errors.Wrap(err, "parse 'conf/app.yaml'")
	}
	return nil
}

func InstallConf(data map[string]string) error {
	err := ReadConf()
	if err != nil {
		return err
	}

	err = renderSection()
	if err != nil {
		return err
	}

	customConf := filepath.Join(CustomDir(), "conf", "app.yaml")

	if !isExist(filepath.Dir(customConf)) {
		err = os.MkdirAll(filepath.Dir(customConf), os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "MkdirAll")
		}
	}

	// Update configuration values
	appConfig.AppName = App.Name
	appConfig.BrandName = App.BrandName
	appConfig.RunUser = App.RunUser
	appConfig.RunMode = "prod"

	// Update log settings
	appConfig.Log.RootPath = Log.RootPath

	// Update web settings
	appConfig.Web.HTTPPort = 9191
	// admin_path := fmt.Sprintf("/uptimepk_%s", randString(6))
	admin_path := "uptimepk"
	appConfig.Web.AdminPath = admin_path

	// Update database settings
	if strings.EqualFold(data["type"], "mysql") {
		appConfig.Database.Type = "mysql"
		appConfig.Database.Hostname = data["hostname"]
		// Convert string port to int64
		hostport, _ := strconv.ParseInt(data["hostport"], 10, 64)
		appConfig.Database.Hostport = hostport
		appConfig.Database.Name = data["dbname"]
		appConfig.Database.User = data["username"]
		appConfig.Database.Password = data["password"]
		appConfig.Database.TablePrefix = data["table_prefix"]
	} else if strings.EqualFold(data["type"], "sqlite3") {
		appConfig.Database.Type = "sqlite3"
		appConfig.Database.Path = data["dbpath"]
	}

	// Update security settings
	appConfig.Security.InstallLock = true
	appConfig.Security.SecretKey = randString(32)

	// Create save config (excludes general and admin)
	saveConfig := YAMLConfigCustom{
		AppName:   appConfig.AppName,
		BrandName: appConfig.BrandName,
		RunUser:   appConfig.RunUser,
		RunMode:   appConfig.RunMode,
		Log:       appConfig.Log,
		Session:   appConfig.Session,
		Web:       appConfig.Web,
		Security:  appConfig.Security,
		Database:  appConfig.Database,
	}

	// Save the updated configuration
	yamlData, err := yaml.Marshal(saveConfig)
	if err != nil {
		return errors.Wrap(err, "marshal yaml config")
	}

	if err := os.WriteFile(customConf, yamlData, os.ModePerm); err != nil {
		return errors.Wrap(err, "write custom config file")
	}

	// write custom configuration file, rewrite initialization read
	err = InitConf("")
	if err != nil {
		return err
	}
	return nil
}

// Init initializes the configuration system
func InitConf(customConf string) error {
	// Load embedded configuration
	data, err := embed.Conf.ReadFile("conf/app.yaml")
	if err != nil {
		return errors.Wrap(err, "read embedded config")
	}

	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		return errors.Wrap(err, "parse 'conf/app.yaml'")
	}

	// Determine custom config path
	if customConf == "" {
		customConf = filepath.Join(CustomDir(), "conf", "app.yaml")
	} else {
		customConf, err = filepath.Abs(customConf)
		if err != nil {
			return errors.Wrap(err, "get absolute path")
		}
	}
	CustomConf = customConf

	// Load custom configuration if exists
	if isFile(customConf) {
		customData, err := os.ReadFile(customConf)
		if err != nil {
			return errors.Wrapf(err, "read custom config %q", customConf)
		}

		// Unmarshal custom config, which will override embedded config
		err = yaml.Unmarshal(customData, &appConfig)
		if err != nil {
			return errors.Wrapf(err, "parse custom config %q", customConf)
		}
	}

	err = renderSection()
	if err != nil {
		return err
	}

	return nil
}

func renderSection() error {
	// Map YAML config to global structs
	App.Name = appConfig.AppName
	App.BrandName = appConfig.BrandName
	App.RunUser = appConfig.RunUser
	App.RunMode = appConfig.RunMode

	// ****************************
	// ----- general settings -----
	// ****************************
	General.MenuFile = appConfig.General.MenuFile

	// ****************************
	// ----- Web settings -----
	// ****************************
	Web.HTTPAddr = appConfig.Web.HTTPAddr
	Web.HTTPPort = appConfig.Web.HTTPPort
	Web.AdminPath = appConfig.Web.AdminPath
	Web.EnableGzip = appConfig.Web.EnableGzip

	// ***************************
	// ----- Log settings -----
	// ***************************
	Log.Format = appConfig.Log.Format
	Log.RootPath = appConfig.Log.RootPath

	// ***************************
	// ----- Database settings -----
	// ***************************
	Database.Type = appConfig.Database.Type
	Database.Path = appConfig.Database.Path
	Database.DSN = appConfig.Database.DSN
	Database.TablePrefix = appConfig.Database.TablePrefix
	Database.Hostname = appConfig.Database.Hostname
	Database.Hostport = appConfig.Database.Hostport
	Database.Name = appConfig.Database.Name
	Database.User = appConfig.Database.User
	Database.Password = appConfig.Database.Password
	Database.SSLMode = appConfig.Database.SSLMode

	// ***************************
	// ----- Security settings -----
	// ***************************
	Security.InstallLock = appConfig.Security.InstallLock
	Security.SecretKey = appConfig.Security.SecretKey
	Security.LoginRememberDays = appConfig.Security.LoginRememberDays
	Security.CookieRememberName = appConfig.Security.CookieRememberName
	Security.CookieUsername = appConfig.Security.CookieUsername
	Security.CookieSecure = appConfig.Security.CookieSecure
	Security.EnableLoginStatusCookie = appConfig.Security.EnableLoginStatusCookie
	Security.LoginStatusCookieName = appConfig.Security.LoginStatusCookieName

	// ***************************
	// ----- Session settings -----
	// ***************************
	Session.Provider = appConfig.Session.Provider
	Session.ProviderConfig = appConfig.Session.ProviderConfig
	Session.CookieName = appConfig.Session.CookieName
	Session.CookieSecure = appConfig.Session.CookieSecure
	Session.GCInterval = appConfig.Session.GCInterval
	Session.MaxLifeTime = appConfig.Session.MaxLifeTime
	Session.CSRFCookieName = appConfig.Session.CSRFCookieName

	return nil
}
