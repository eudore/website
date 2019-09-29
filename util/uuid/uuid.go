package uuid

type Uuid interface {
	GetId() string
}

var defauleId Uuid

func GetId() string {
	return defauleId.GetId()
}

func init() {
	defauleId, _ = NewWorker(0)
}
