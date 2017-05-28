package moduleserver

/*

type supManager struct {
	work       chan chan childChan
	supervisor []*Supervisor
	chanManager
}
type chanManager struct {
	children map[chan childChan]chan childChan
}

type childChan int

*/
/*func newSupManager() *supManager {
	supManager = new(supManager)
	supManager.supervisor = make([]*Supervisor, 0)
	return supManager

}*/ /*


func (s *supManager) init() {
	s.supervisor = make([]*Supervisor, 0)
	s.work = make(chan int)
	s.chanManager = new(chanManager)
	s.chanManager.children = make(map[string]chan childChan)

}

// must be called once
func Start() {
	supManager.init()
	defer restart(startServerManager)
	startServerManager()
}

func startServerManager() {
	for {

		data := <-supManager.work
		if data == nil {
			StartSupervisor()
		} else {
			panic("Unknown data")
		}

	}
}

func restart(module func()) {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic")
		module()
	}
}
*/
