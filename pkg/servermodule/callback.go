package servermodule

type callback interface {
	handleCast()
	handleCall()
	init()
	terminate()
}
