package core

type Crosslink struct {
	Shard      Shard
	ParentRoot Root
	// Crosslinking data
	StartEpoch Epoch
	EndEpoch   Epoch
	DataRoot   Root
}
