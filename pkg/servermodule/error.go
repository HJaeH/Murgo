package servermodule

//todo : 진행 하면서 필요한 것 추가/
const (
	InvalidModuleID = 100 + iota
)

const ErrorMap = map[int]string{
	InvalidModuleID: "Module ID is not registerd",
}
