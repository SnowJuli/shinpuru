package inits

import (
	"strings"

	"github.com/sarulabs/di/v2"
	"github.com/zekroTJA/shinpuru/internal/config"
	"github.com/zekroTJA/shinpuru/internal/services/database"
	"github.com/zekroTJA/shinpuru/internal/services/database/mysql"
	"github.com/zekroTJA/shinpuru/internal/services/database/redis"
	"github.com/zekroTJA/shinpuru/internal/services/database/sqlite"
	"github.com/zekroTJA/shinpuru/internal/util"
	"github.com/zekroTJA/shinpuru/internal/util/static"
)

func InitDatabase(container di.Container) database.Database {
	var db database.Database
	var err error

	cfg := container.Get(static.DiConfig).(*config.Config)

	switch strings.ToLower(cfg.Database.Type) {
	case "mysql", "mariadb":
		db = new(mysql.MysqlMiddleware)
		err = db.Connect(cfg.Database.MySql)
	case "sqlite", "sqlite3":
		db = new(sqlite.SqliteMiddleware)
		err = db.Connect(cfg.Database.Sqlite)
		printSqliteWraning()
	}

	if m, ok := db.(database.Migration); ok {
		util.Log.Info("Checking database for migrations and apply if needed...")
		if err = m.Migrate(); err != nil {
			util.Log.Fatal("Database migration failed:", err)
		}
	} else {
		util.Log.Warning("Skip database migration: middleware does not support migrations")
	}

	if cfg.Database.Redis != nil && cfg.Database.Redis.Enable {
		db = redis.NewRedisMiddleware(cfg.Database.Redis, db)
		util.Log.Info("Enabled redis as database cache")
	}

	if err != nil {
		util.Log.Fatal("Failed connecting to database:", err)
	}
	util.Log.Info("Connected to database")

	return db
}

func printSqliteWraning() {
	util.Log.Warning("--------------------------[ ATTENTION ]--------------------------")
	util.Log.Warning("You are currently using SQLite as database driver. Please ONLY   ")
	util.Log.Warning("use SQLite during testing and debugging and NEVER use SQLite in a")
	util.Log.Warning("real production environment! Here you can read about why:        ")
	util.Log.Warning("https://github.com/zekroTJA/shinpuru/wiki/No-SQLite-in-production")
	util.Log.Warning("-----------------------------------------------------------------")
}
