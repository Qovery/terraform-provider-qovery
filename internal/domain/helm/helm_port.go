package helm

const (
	DefaultProtocol = ProtocolHttp
)

type Port struct {
	Name         string
	InternalPort int32
	ExternalPort *int32
	ServiceName  string
	Namespace    *string
	Protocol     Protocol
	IsDefault    bool
}
