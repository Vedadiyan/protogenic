package global

import "sync"

type DEPENDENCY_TYPES string

const (
	POSTGRES DEPENDENCY_TYPES = "postgres"
	REDIS    DEPENDENCY_TYPES = "redis"
	MONGO    DEPENDENCY_TYPES = "mongo"
	NATS     DEPENDENCY_TYPES = "nats"
	INFLUX   DEPENDENCY_TYPES = "influx"
	DGRAPH   DEPENDENCY_TYPES = "dgraph"
)

var (
	_mut          sync.RWMutex
	_dependencies map[DEPENDENCY_TYPES]map[string]bool
)

func init() {
	_dependencies = make(map[DEPENDENCY_TYPES]map[string]bool)
}

func Register(dependencyType DEPENDENCY_TYPES, key string) {
	// _mut.RLock()
	// defer _mut.Unlock()
	dt, ok := _dependencies[dependencyType]
	if !ok {
		//_mut.Lock()
		dt = make(map[string]bool)
		//_mut.Unlock()
	}
	_, ok = dt[key]
	if !ok {
		//_mut.Lock()
		dt[key] = true
		_dependencies[dependencyType] = dt
		//_mut.Unlock()
	}
}

func ForEach(callback func(dependencyType DEPENDENCY_TYPES, key string)) {
	//_mut.RLock()
	//defer _mut.RUnlock()
	for dt, value := range _dependencies {
		for key := range value {
			callback(dt, key)
		}
	}
}
