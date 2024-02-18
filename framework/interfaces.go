package framework

type Checker[SourceType any, VersionType any] interface {
	Check(CheckRequest[SourceType, VersionType]) ([]VersionType, error)
}

type Getter[SourceType any, VersionType any, ParamsType any] interface {
	Get(baseDir string, request GetRequest[SourceType, VersionType, ParamsType]) (Response[VersionType], error)
}

type Putter[SourceType any, VersionType any, ParamsType any] interface {
	Put(baseDir string, request PutRequest[SourceType, ParamsType]) (Response[VersionType], error)
}
