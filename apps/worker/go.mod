module github.com/vandordev/vkit-zenith/apps/worker

go 1.27.0

require (
	github.com/vandordev/vkit-zenith/packages/config/go/config v0.0.0
	github.com/vandordev/vkit-zenith/packages/db v0.0.0
	github.com/riverqueue/river v0.51.2
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.51.2
)

require (
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/riverqueue/river/riverdriver v0.51.2 // indirect
	github.com/riverqueue/river/rivershared v0.51.2 // indirect
	github.com/riverqueue/river/rivertype v0.51.2 // indirect
	github.com/tidwall/gjson v1.20.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/vandordev/vkit-zenith/packages/config/go/config => ../../packages/config/go

replace github.com/vandordev/vkit-zenith/packages/db => ../../packages/db
