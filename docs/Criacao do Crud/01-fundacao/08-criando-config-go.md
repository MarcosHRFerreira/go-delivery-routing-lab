# Criando O `config.go`

Este documento explica como montar a base de configuracao da aplicacao.

O objetivo do `config.go` e centralizar:

- leitura de variaveis de ambiente
- aplicacao de valores padrao
- validacao de campos obrigatorios
- helpers pequenos para o restante do projeto

---

## Onde Criar

Digite em:

- `internal/config/config.go`

---

## Responsabilidade Do Arquivo

Esse arquivo nao deve abrir conexao com banco, nao deve subir servidor e nao deve conhecer regra de negocio.

Ele existe para:

- montar um objeto `Config`
- carregar `.env` em ambiente local
- validar se as variaveis obrigatorias existem

---

## Estrutura Sugerida

```go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBUrlMigration string
	SecretJwt      string

	DBHost     string
	DBUser     string
	DBName     string
	DBPassword string
	DBPort     string
}
```

### Leitura Da Estrutura

- `Port` define onde a API sobe
- `DBUrlMigration` pode ser usada por ferramentas como `dbmate`
- `SecretJwt` fica pronta para autenticacao futura
- `DBHost`, `DBUser`, `DBName`, `DBPassword` e `DBPort` alimentam a conexao com MySQL

---

## Funcao Principal

Digite em `internal/config/config.go`:

```go
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		DBUrlMigration: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SecretJwt:      strings.TrimSpace(os.Getenv("SECRET_JWT")),
		DBHost:         strings.TrimSpace(os.Getenv("DB_HOST")),
		DBUser:         strings.TrimSpace(os.Getenv("DB_USER")),
		DBName:         strings.TrimSpace(os.Getenv("DB_NAME")),
		DBPassword:     strings.TrimSpace(os.Getenv("DB_PASSWORD")),
		DBPort:         getEnvOrDefault("DB_PORT", "3306"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
```

### Por Que Esse Fluxo E Bom

- o `.env` e carregado so no bootstrap
- `strings.TrimSpace(...)` evita lixo vindo do ambiente
- `getEnvOrDefault(...)` resolve defaults simples
- `validate()` protege o sistema de subir com configuracao quebrada

---

## Helpers Do Arquivo

Digite tambem:

```go
func (c *Config) ServerAddress() string {
	return ":" + c.Port
}

func (c *Config) validate() error {
	missingFields := make([]string, 0, 5)

	if c.SecretJwt == "" {
		missingFields = append(missingFields, "SECRET_JWT")
	}
	if c.DBHost == "" {
		missingFields = append(missingFields, "DB_HOST")
	}
	if c.DBUser == "" {
		missingFields = append(missingFields, "DB_USER")
	}
	if c.DBName == "" {
		missingFields = append(missingFields, "DB_NAME")
	}
	if c.DBPassword == "" {
		missingFields = append(missingFields, "DB_PASSWORD")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}
```

---

## Variaveis Esperadas

Para esse projeto, o `config.go` espera algo proximo de:

```env
PORT=8080
DATABASE_URL=mysql://dbeaver:superSecret@127.0.0.1:3306/delivery_routing_lab
SECRET_JWT=dev-secret
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=dbeaver
DB_PASSWORD=superSecret
DB_NAME=delivery_routing_lab
```

---

## Boas Praticas

- nao espalhar `os.Getenv(...)` pelo projeto
- nao retornar `panic` ao faltar configuracao
- centralizar os nomes das variaveis nesse pacote
- deixar defaults apenas para campos realmente seguros

---

## Proximo Documento

Depois de `config.go`, o proximo passo natural e criar:

- `internal/apperror/error.go`

Esse arquivo padroniza como o sistema representa erros de negocio e erros HTTP.
