package devices

import (
	"simpletcp/src/tcpserver"
	"sync"
)

type Store interface {
	Add(deviceID string, conn tcpserver.TcpConn)
	Delete(deviceID string)
	GetByDeviceID(deviceID string) tcpserver.TcpConn
	List() []tcpserver.Property
}

var DeviceStore Store

func init() {
	DeviceStore = NewDeviceStore()
}

type deviceStore struct {
	deviceConn *sync.Map
}

func NewDeviceStore() Store {
	return &deviceStore{
		deviceConn: new(sync.Map),
	}
}

func (ds *deviceStore) Add(deviceID string, conn tcpserver.TcpConn) {
	ds.deviceConn.Store(deviceID, conn)
}

func (ds *deviceStore) Delete(deviceID string) {
	ds.deviceConn.Delete(deviceID)
}

func (ds *deviceStore) GetByDeviceID(deviceID string) tcpserver.TcpConn {
	conn, ok := ds.deviceConn.Load(deviceID)
	if ok {
		return conn.(tcpserver.TcpConn)
	} else {
		return nil
	}
}

func (ds *deviceStore) List() []tcpserver.Property {
	ret := make([]tcpserver.Property, 0)
	ds.deviceConn.Range(func(key, value interface{}) bool {
		conn := value.(tcpserver.TcpConn)
		ret = append(ret, conn.GetProperty())
		return true
	})
	return ret
}
