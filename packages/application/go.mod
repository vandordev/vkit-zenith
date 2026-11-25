module github.com/vandordev/vkit-zenith/packages/application

go 1.27.0

require (
	github.com/jackc/pgx/v5 v5.11.0
	github.com/nyaruka/phonenumbers v1.8.1
	github.com/vandordev/vkit-zenith/packages/db v0.0.0
	golang.org/x/crypto v0.58.0
)

replace github.com/vandordev/vkit-zenith/packages/db => ../db

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
