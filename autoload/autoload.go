package autoload

import (
	"github.com/Valgard/godotenv"
)

func init() {
	dotEnv := godotenv.New()
	_ = dotEnv.LoadEnv(".env")
}
