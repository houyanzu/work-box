package modulelock

import (
	"sync"
	"time"
)

type ModuleManager struct {
	ModuleLocks sync.Map
}

var (
	um   *ModuleManager
	once sync.Once
)

func initModuleManager() {
	um = new(ModuleManager)
}

func GetInstance() *ModuleManager {
	once.Do(initModuleManager)
	return um
}

func (um *ModuleManager) Lock(module string, timeout time.Duration) {
	lock, _ := um.ModuleLocks.LoadOrStore(module, new(sync.Mutex))
	lock.(*sync.Mutex).Lock()

	// 在锁定的锁超时后解锁并删除
	if timeout > 0 {
		go func() {
			<-time.After(timeout)
			um.ModuleLocks.Delete(module)
		}()
	}
}

func (um *ModuleManager) Unlock(module string) {
	lock, loaded := um.ModuleLocks.Load(module)
	if loaded {
		lock.(*sync.Mutex).Unlock()
	}
}
