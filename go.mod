module github.com/yttydcs/myflowhub-server

go 1.23.0

toolchain go1.24.5

require (
	github.com/jackc/pgx/v5 v5.7.6
	github.com/yttydcs/myflowhub-core v0.1.0
	github.com/yttydcs/myflowhub-proto v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/yttydcs/myflowhub-core => ../MyFlowHub-Core

replace github.com/yttydcs/myflowhub-proto => ../MyFlowHub-Proto
