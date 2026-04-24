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

var yamlConfig YAMLConfig

func ReadConf() error {
	data, err := embed.Conf.ReadFile("conf/app.yaml")
	if err != nil {
		return errors.Wrap(err, "read file 'conf/app.yaml'")
	}

	err = yaml.Unmarshal(data, &yamlConfig)
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
	yamlConfig.AppName = App.Name
	yamlConfig.BrandName = App.BrandName
	yamlConfig.RunUser = App.RunUser
	yamlConfig.RunMode = "prod"

	// Update log settings
	yamlConfig.Log.RootPath = Log.RootPath

	// Update web settings
	yamlConfig.Web.HTTPPort = 9191
	// admin_path := fmt.Sprintf("/mgo_%s", randString(6))
	admin_path := "uptimepk"
	yamlConfig.Web.AdminPath = admin_path

	// Update database settings
	if strings.EqualFold(data["type"], "mysql") {
		yamlConfig.Database.Type = "mysql"
		yamlConfig.Database.Hostname = data["hostname"]
		// Convert string port to int64
		hostport, _ := strconv.ParseInt(data["hostport"], 10, 64)
		yamlConfig.Database.Hostport = hostport
		yamlConfig.Database.Name = data["dbname"]
		yamlConfig.Database.User = data["username"]
		yamlConfig.Database.Password = data["password"]
		yamlConfig.Database.TablePrefix = data["table_prefix"]
	} else if strings.EqualFold(data["type"], "sqlite3") {
		yamlConfig.Database.Type = "sqlite3"
		yamlConfig.Database.Path = data["dbpath"]
	}

	// Update security settings
	yamlConfig.Security.InstallLock = true
	yamlConfig.Security.SecretKey = randString(32)

	// Create save config (excludes general and admin)
	saveConfig := YAMLConfigCustom{
		AppName:   yamlConfig.AppName,
		BrandName: yamlConfig.BrandName,
		RunUser:   yamlConfig.RunUser,
		RunMode:   yamlConfig.RunMode,
		Log:       yamlConfig.Log,
		Session:   yamlConfig.Session,
		Web:       yamlConfig.Web,
		Security:  yamlConfig.Security,
		Database:  yamlConfig.Database,
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

	err = yaml.Unmarshal(data, &yamlConfig)
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
		err = yaml.Unmarshal(customData, &yamlConfig)
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
	App.Name = yamlConfig.AppName
	App.BrandName = yamlConfig.BrandName
	App.RunUser = yamlConfig.RunUser
	App.RunMode = yamlConfig.RunMode

	// ****************************
	// ----- general settings -----
	// ****************************
	General.MenuFile = yamlConfig.General.MenuFile

	// ****************************
	// ----- Web settings -----
	// ****************************
	Web.HTTPAddr = yamlConfig.Web.HTTPAddr
	Web.HTTPPort = yamlConfig.Web.HTTPPort
	Web.AdminPath = yamlConfig.Web.AdminPath
	Web.EnableGzip = yamlConfig.Web.EnableGzip

	// ***************************
	// ----- Log settings -----
	// ***************************
	Log.Format = yamlConfig.Log.Format
	Log.RootPath = yamlConfig.Log.RootPath

	// ***************************
	// ----- Database settings -----
	// ***************************
	Database.Type = yamlConfig.Database.Type
	Database.Path = yamlConfig.Database.Path
	Database.DSN = yamlConfig.Database.DSN
	Database.TablePrefix = yamlConfig.Database.TablePrefix
	Database.Hostname = yamlConfig.Database.Hostname
	Database.Hostport = yamlConfig.Database.Hostport
	Database.Name = yamlConfig.Database.Name
	Database.User = yamlConfig.Database.User
	Database.Password = yamlConfig.Database.Password
	Database.SSLMode = yamlConfig.Database.SSLMode

	// ***************************
	// ----- Security settings -----
	// ***************************
	Security.InstallLock = yamlConfig.Security.InstallLock
	Security.SecretKey = yamlConfig.Security.SecretKey
	Security.LoginRememberDays = yamlConfig.Security.LoginRememberDays
	Security.CookieRememberName = yamlConfig.Security.CookieRememberName
	Security.CookieUsername = yamlConfig.Security.CookieUsername
	Security.CookieSecure = yamlConfig.Security.CookieSecure
	Security.EnableLoginStatusCookie = yamlConfig.Security.EnableLoginStatusCookie
	Security.LoginStatusCookieName = yamlConfig.Security.LoginStatusCookieName

	// ***************************
	// ----- Session settings -----
	// ***************************
	Session.Provider = yamlConfig.Session.Provider
	Session.ProviderConfig = yamlConfig.Session.ProviderConfig
	Session.CookieName = yamlConfig.Session.CookieName
	Session.CookieSecure = yamlConfig.Session.CookieSecure
	Session.GCInterval = yamlConfig.Session.GCInterval
	Session.MaxLifeTime = yamlConfig.Session.MaxLifeTime
	Session.CSRFCookieName = yamlConfig.Session.CSRFCookieName

	return nil
}
