module github.com/vandordev/vkit-zenith/apps/api

go 1.27.0

require (
	github.com/danielgtaylor/huma/v2 v2.39.1
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.16.2
	github.com/vandordev/vkit-zenith/packages/config/go/config v0.0.0
	github.com/vandordev/vkit-zenith/packages/fixtures v0.0.0
	go.uber.org/fx v1.24.0
)

require (
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/labstack/echo/v5 v5.3.0 // indirect
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.23 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

replace github.com/vandordev/vkit-zenith/packages/config/go/config => ../../packages/config/go

replace github.com/vandordev/vkit-zenith/packages/fixtures => ../../packages/fixtures
