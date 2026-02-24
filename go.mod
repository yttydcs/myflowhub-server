module github.com/yttydcs/myflowhub-server

go 1.23.0

toolchain go1.24.5

require (
	github.com/yttydcs/myflowhub-core v0.2.1
	github.com/yttydcs/myflowhub-proto v0.1.1
	github.com/yttydcs/myflowhub-subproto/auth v0.1.0
	github.com/yttydcs/myflowhub-subproto/exec v0.1.0
	github.com/yttydcs/myflowhub-subproto/file v0.1.0
	github.com/yttydcs/myflowhub-subproto/flow v0.1.0
	github.com/yttydcs/myflowhub-subproto/forward v0.1.0
	github.com/yttydcs/myflowhub-subproto/management v0.1.1
	github.com/yttydcs/myflowhub-subproto/topicbus v0.1.0
	github.com/yttydcs/myflowhub-subproto/varstore v0.1.0
)

require github.com/yttydcs/myflowhub-subproto/broker v0.1.0 // indirect
