# godotenv

This module parses `.env` files to make environment variables stored in them accessible via `os.Getenv("key")`

> This is a port of the Symfony 
> [Dotenv](https://symfony.com/doc/current/components/dotenv.html "Homepage of the Symfony Dotenv Component") Component 

## Installation

```shell script
go get https://github.com/Valgard/godotenv 
```

## Usage

Sensitive information and environment-dependent settings should be defined as environment variables (as recommended for 
[twelve-factor](https://12factor.net/) applications). Using a `.env` file to store those environment variables eases 
development and CI management by keeping them in one "standard" place and agnostic of the technology stack you are 
using.

Load a `.env` file in your application via Dotenv::Load():

```go
package main

import (
	"github.com/Valgard/godotenv"
)

func main() {
	dotenv := godotenv.New()
	if err := dotenv.Load(".env"); err != nil {
		panic(err)
	}

	// You can also load several files
	if err := dotenv.Load(".env", ".env.dev"); err != nil {
		panic(err)
	}
}
```

Given the following `.env` file content:

```.env
# .env
DB_USER=root
DB_PASS=pass
```

Access the value with os.Getenv() in your code:

```go
dbUser := os.Getenv("DB_USER")
```

The `Load()` method never overwrites existing environment variables. Use the `Overload()` method if you need to 
overwrite them:

```go
if err := dotenv.Overload(".env"); err != nil {
	panic(err)
}
```

As you're working with this module you'll notice that you might want to have different files depending on the
environment you're working in. Typically this happens for local development or Continuous Integration where you might
want to have different files for your `test` and `dev` environments.

You can use Dotenv::LoadEnv() to ease this process:

```go
	dotenv := godotenv.New()
	if err := dotenv.LoadEnv(".env"); err != nil {
		panic(err)
	}
```

The Dotenv module will then look for the correct `.env` file to load in the following order whereas the files loaded 
later override the variables defined in previously loaded files:

1. If `.env` exists, it is loaded first. In case there's no `.env` file but a `.env.dist`, this one will be 
   loaded instead.
2. If one of the previously mentioned files contains the `APP_ENV` variable, the variable is populated and used 
   to load environment-specific files hereafter. If `APP_ENV` is not defined in either of the previously mentioned
   files, `dev` is assumed for `APP_ENV` and populated by default.
3. If there's a `.env.local` representing general local environment variables it's loaded now.
4. If there's a `.env.{env}.local` file, this one is loaded. Otherwise, it falls back to `.env.{env}`.

This might look complicated at first glance but it gives you the opportunity to commit multiple environment-specific
files that can then be adjusted to your local environment. Given you commit `.env`, `.env.test` and `.env.dev` to represent
different configuration settings for your environments, each of them can be adjusted by using `.env.local`,
`.env.test.local` and `.env.dev.local` respectively.

> `.env.local` is always ignored in `test` environment because tests should produce the same results for everyone.

You can adjust the variable defining the environment, default environment and test environments by passing them as 
additional options to Dotenv::LoadEnv() (see LoadEnv() for details).

You should never store a `.env.local` file in your code repository as it might contain sensitive information;
create a `.env` file (or multiple environment-specific ones as shown above) with sensible defaults instead.

> This dotenv module can be used in any environment of your application: development, testing, staging and 
> even production. However, in production it's recommended to configure real environment variables to avoid 
> the performance overhead of parsing the .env file for every request.

As a `.env` file is a regular shell script, you can source it in your own shell scripts:

```shell script
source .env
```

Add comments by prefixing them with `#`:
```.env
# Database credentials
DB_USER=root
DB_PASS=pass # This is the secret password
```

Use environment variables in values by prefixing variables with `$`:

```.env
DB_USER=root
DB_PASS=${DB_USER}pass # Include the user as a password prefix
```

> The order is important when some env var depends on the value of other env vars. In the above example, `DB_PASS`
> must be defined after `DB_USER`. Moreover, if you define multiple `.env` files and put `DB_PASS` first, its value will 
> depend on the `DB_USER` value defined in other files instead of the value defined in this file.

Define a default value in case the environment variable is not set:

```.env
DB_USER=
DB_PASS=${DB_USER:-root}pass # results in DB_PASS=rootpass
```

Embed commands via `$()` (not supported on Windows):

```.env
START_TIME=$(date)
```

> Note that using `$()` might not work depending on your shell.


*This work, including the code samples, is licensed under a Creative Commons BY-SA 3.0 license.*
