package godotenv

var (
	// standard dotEnv instance
	dotenv *dotEnv
)

// Sets dotEnv options and returns previous option values
func Option(opts ...option) (previous []option) { return dotenv.Option(opts...) }

// Loads one or several .env files.
func Load(path string, extraPaths ...string) error { return dotenv.Load(path, extraPaths...) }

// Loads a .env file and the corresponding .env.local, .env.{env} and .env.{env}.local files if they exist.
//
// .env.local is always ignored in test env because tests should produce the same results for everyone.
//
// .env.dist is loaded when it exists and .env is not found.
func LoadEnv(path string, opts ...option) error { return dotenv.LoadEnv(path, opts...) }

// Loads one or several .env files and enables override existing vars.
func Overload(path string, extraPaths ...string) error { return dotenv.Overload(path, extraPaths...)}

// Sets values as environment variables.
func Populate(values map[string]string, overrideExistingVars bool) error { return dotenv.Populate(values, overrideExistingVars) }

// Parses the contents of an .env file
func Parse(data string, path string) (map[string]string, error) { return dotenv.Parse(data, path) }
