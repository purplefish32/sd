package watchers

type ConnectHandler func(instanceID string, deviceID string, productID uint16) error
type DisconnectHandler func(instanceID string, deviceID string) error
