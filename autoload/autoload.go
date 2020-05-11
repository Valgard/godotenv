package autoload

import (
	dotEnv "github.com/Valgard/godotenv"
)

func init() {
	_ = dotEnv.LoadEnv(".env")
}
