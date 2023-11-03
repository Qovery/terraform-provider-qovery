package gittoken

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type GitTokenType string

const (
	GITHUB    GitTokenType = "GITHUB"
	GITLAB    GitTokenType = "GITLAB"
	BITBUCKET GitTokenType = "BITBUCKET"
)

var AllowedGitTokenTypeValues = []GitTokenType{GITHUB, GITLAB, BITBUCKET}

func (v GitTokenType) String() string {
	return string(v)
}

func (v GitTokenType) Validate() error {
	if slices.Contains(AllowedGitTokenTypeValues, v) {
		return nil
	}

	return fmt.Errorf("invalid value '%v' for Kind: valid values are %v", v, AllowedGitTokenTypeValues)
}

func NewGitTokenTypeFromString(v string) (*GitTokenType, error) {
	ev := GitTokenType(v)

	if err := ev.Validate(); err != nil {
		return nil, err
	}

	return &ev, nil
}
