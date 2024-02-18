package framework

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CheckRequest[S any, V any] struct {
	Source  S  `json:"source"`
	Version *V `json:"version"`
}

type GetRequest[S any, V any, P any] struct {
	Source  S `json:"source"`
	Version V `json:"version"`
	Params  P `json:"params"`
}

type PutRequest[S any, P any] struct {
	Source S `json:"source"`
	Params P `json:"params"`
}

type Response[V any] struct {
	Version  V               `json:"version"`
	Metadata []MetadataField `json:"metadata"`
}
