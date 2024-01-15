package helm

type RawValue struct {
	Name    string
	Content string
}

type Raw struct {
	Values []RawValue
}

type ValuesOverrideGit struct {
	Url      string
	Branch   string
	Paths    []string
	GitToken *string
}

type ValuesOverrideFile struct {
	Raw           *Raw
	GitRepository *ValuesOverrideGit
}

type ValuesOverride struct {
	Set       [][]string
	SetString [][]string
	SetJson   [][]string
	File      *ValuesOverrideFile
}

type NewHelmValuesOverrideParams struct {
	Set       [][]string
	SetString [][]string
	SetJson   [][]string
	File      *ValuesOverrideFile
}

type NewHelmPortParams struct {
	Name         string
	InternalPort int32
	ExternalPort *int32
	ServiceName  string
	Namespace    *string
	Protocol     string
	IsDefault    bool
}

func NewHelmValuesOverride(params NewHelmValuesOverrideParams) (*ValuesOverride, error) {
	newValuesOverride := &ValuesOverride{
		Set:       params.Set,
		SetString: params.SetString,
		SetJson:   params.SetJson,
		File:      params.File,
	}

	return newValuesOverride, nil
}
