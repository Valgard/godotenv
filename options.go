package godotenv

type option func(*dotEnv) option

// Sets a list of app envs for which sets the debug status to false
func ProdEnvs(prodEnvs []string) option {
	return func(d *dotEnv) option {
		previous := d.prodEnvs
		d.prodEnvs = prodEnvs

		return ProdEnvs(previous)
	}
}

// Sets a list of app envs for which .env.local should be ignored
func TestEnvs(testEnvs []string) option {
	return func(d *dotEnv) option {
		previous := d.testEnvs
		d.testEnvs = testEnvs

		return TestEnvs(previous)
	}
}

// Sets the name of the env vars that defines the app env
func EnvKey(envKey string) option {
	return func(d *dotEnv) option {
		previous := d.envKey
		d.envKey = envKey

		return EnvKey(previous)
	}
}

// Sets the name of the env vars that defines the app debug status
func DebugKey(debugKey string) option {
	return func(d *dotEnv) option {
		previous := d.debugKey
		d.envKey = debugKey

		return DebugKey(previous)
	}
}

// Sets the app env to use when none is defined
func DefaultEnv(defaultEnv string) option {
	return func(d *dotEnv) option {
		previous := d.defaultEnv
		d.defaultEnv = defaultEnv

		return DefaultEnv(previous)
	}
}
