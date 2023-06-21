package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type IConfig interface {
	App() IAppConfig
	Db() IDbConfig
	Jwt() IJwtConfig
}
type config struct {
	app *app
	db  *db
	jwt *jwt
}

func NewConfig(path string) IConfig {
	if err := godotenv.Load(path); err != nil {
		log.Fatalf("load dotenv failed: %v", err)
	}
	return &config{
		app: &app{
			token: os.Getenv("APP_TOKEN"),
		},
		db: &db{
			host:     os.Getenv("MONGODB_HOST"),
			port:     os.Getenv("MONGODB_PORT"),
			dbname:   os.Getenv("MONGODB_DBNAME"),
			username: os.Getenv("MONGODB_USERNAME"),
			password: os.Getenv("MONGODB_PASSWORD"),
		},
		jwt: &jwt{
			secertKey: os.Getenv("JWT_SECRET"),
		},
	}
}

// App
type IAppConfig interface {
	GetToken() string
}
type app struct {
	token string
}

func (cfg *config) App() IAppConfig {
	return cfg.app
}
func (a *app) GetToken() string {
	return a.token
}

// Database
type IDbConfig interface {
	Url() string
	Dbname() string
}
type db struct {
	url      string
	username string
	password string
	host     string
	port     string
	dbname   string
}

func (cfg *config) Db() IDbConfig {
	return cfg.db
}
func (d *db) Url() string {
	return fmt.Sprintf(
		"mongodb+srv://%s:%s@%s%s",
		d.username,
		d.password,
		d.host,
		d.port,
	)
}
func (d *db) Dbname() string {
	return d.dbname
}

type IJwtConfig interface {
	SecretKey() []byte
}
type jwt struct {
	secertKey string
}

func (c *config) Jwt() IJwtConfig {
	return c.jwt
}
func (j *jwt) SecretKey() []byte { return []byte(j.secertKey) }
