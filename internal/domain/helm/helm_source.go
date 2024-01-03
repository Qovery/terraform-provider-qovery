package helm

type SourceGitRepository struct {
	Url        string
	Branch     *string
	RootPath   string
	GitTokenId *string
}

type SourceHelmRepository struct {
	RepositoryId string
	ChartName    string
	ChartVersion string
}

type Source struct {
	GitRepository  *SourceGitRepository
	HelmRepository *SourceHelmRepository
}

type NewHelmSourceGitRepository struct {
	Url        string
	Branch     *string
	RootPath   string
	GitTokenId *string
}

type NewHelmSourceHelmRepository struct {
	RepositoryId string
	ChartName    string
	ChartVersion string
}

type NewHelmSourceParams struct {
	HelmSourceGitRepository  *NewHelmSourceGitRepository
	HelmSourceHelmRepository *NewHelmSourceHelmRepository
}

func NewHelmSource(params NewHelmSourceParams) (*Source, error) {
	var gitRepository *SourceGitRepository = nil
	if params.HelmSourceGitRepository != nil {
		gitRepository = &SourceGitRepository{
			Url:        params.HelmSourceGitRepository.Url,
			Branch:     params.HelmSourceGitRepository.Branch,
			RootPath:   params.HelmSourceGitRepository.RootPath,
			GitTokenId: params.HelmSourceGitRepository.GitTokenId,
		}
	}

	var helmRepository *SourceHelmRepository = nil
	if params.HelmSourceHelmRepository != nil {
		helmRepository = &SourceHelmRepository{
			RepositoryId: params.HelmSourceHelmRepository.RepositoryId,
			ChartName:    params.HelmSourceHelmRepository.ChartName,
			ChartVersion: params.HelmSourceHelmRepository.ChartVersion,
		}
	}

	newValuesOverride := &Source{
		GitRepository:  gitRepository,
		HelmRepository: helmRepository,
	}

	return newValuesOverride, nil
}
