package config

type PGXConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MGConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	Port       string
	DBName     string
	DBAuth     string
}
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}
