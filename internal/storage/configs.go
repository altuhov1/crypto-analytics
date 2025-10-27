package storage

type PGXConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MGConfig struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     int
	DBName     string
}
