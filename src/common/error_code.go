package common

type ReturnCode int32

const (
	ReturnCode_ReturnUnknown              ReturnCode = -1
	ReturnCode_OK                         ReturnCode = 0
	ReturnCode_StorageEngineInternalError ReturnCode = 1
	ReturnCode_StorageEngineSubmitFailed  ReturnCode = 2
)

var (
	ReturnCode_name = map[int32]string{
		-1: "ReturnCode_ReturnUnknown",
		0:  "ReturnCode_OK",
		1:  "ReturnCode_StorageEngineInternalError",
		2:  "ReturnCode_StorageEngineSubmitFailed",
	}
)

func (rc ReturnCode) Number() int32 {
	return int32(rc)
}

func (rc ReturnCode) String() string {
	return ReturnCode_name[int32(rc)]
}
