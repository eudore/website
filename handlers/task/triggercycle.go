package task

type (
	CycleTrigger struct {
		Node     *cycleNode
		Schedule chan int
	}
	cycleNode struct {
		Id      int
		Current uint64
		Add     uint64
		Next    *cycleNode
	}
)
