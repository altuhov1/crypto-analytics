package config

type PGXConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MGConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
	DBAuth     string
}
