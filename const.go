package go_eureka_client

type instanceStatus string
type action string
type HaStrategy int

const (
	INSTANCE_STATUS_UP             instanceStatus = "UP"
	INSTANCE_STATUS_DOWN           instanceStatus = "DOWN"
	INSTANCE_STATUS_STARTING       instanceStatus = "STARTING"
	INSTANCE_STATUS_OUT_OF_SERVICE instanceStatus = "OUT_OF_SERVICE"
	INSTANCE_STATUS_UNKNOWN        instanceStatus = "UNKNOWN"

	ACTION_TYPE_ADDED    action = "ADDED"
	ACTION_TYPE_MODIFIED action = "MODIFIED"
	ACTION_TYPE_DELETED  action = "DELETED"

	HA_STRATEGY_RANDOM HaStrategy = 1
	HA_STRATEGY_STICK  HaStrategy = 2
	HA_STRATEGY_OTHER  HaStrategy = 3

	_DEFAULT_TIME_OUT = 5

	_DEFAULT_EUREKA_SERVER_URL = "http://127.0.0.1:8761/eureka/"

	_DEFAULT_INSTNACE_PORT          = 9090
	_DEFAULT_INSTNACE_SECURE_PORT   = 9443
	_RENEWAL_INTERVAL_IN_SECS       = 30
	_DURATION_IN_SECS               = 90
	_DEFAULT_DATA_CENTER_INFO       = "MyOwn"
	_DEFAULT_DATA_CENTER_INFO_CLASS = "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo"
)

func RenewalIntervalInSecs(s int64) Option {
	return func(instance *Instance) {
		instance.LeaseInfo.RenewalIntervalInSecs = s
	}
}

func InstanceID(InstanceID string) Option {
	return func(instance *Instance) {
		instance.InstanceID = InstanceID
	}
}

func AppName(app string) Option {
	return func(instance *Instance) {
		instance.App = app
	}
}
