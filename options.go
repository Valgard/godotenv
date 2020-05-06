package godotenv

type option func(*dotEnv) option

func ProdEnvs(prodEnvs []string) option {
	return func(d *dotEnv) option {
		previous := d.prodEnvs
		d.prodEnvs = prodEnvs

		return ProdEnvs(previous)
	}
}

func TestEnvs(testEnvs []string) option {
	return func(d *dotEnv) option {
		previous := d.testEnvs
		d.testEnvs = testEnvs

		return TestEnvs(previous)
	}
}

func EnvKey(envKey string) option {
	return func(d *dotEnv) option {
		previous := d.envKey
		d.envKey = envKey

		return EnvKey(previous)
	}
}

func DefaultEnv(defaultEnv string) option {
	return func(d *dotEnv) option {
		previous := d.defaultEnv
		d.defaultEnv = defaultEnv

		return DefaultEnv(previous)
	}
}

