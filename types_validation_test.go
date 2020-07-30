/*
Copyright 2020 The Flux CD contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitprovider

import (
	"fmt"
	"testing"

	"github.com/fluxcd/go-git-providers/validation"
)

type validateMethod string

const (
	validateCreate = validateMethod("Create")
	validateUpdate = validateMethod("Update")
	validateDelete = validateMethod("Delete")
)

type validateFunc func() error

func assertValidation(t *testing.T, structName string, method validateMethod, validateFn validateFunc, expectedErrs []error) {
	funcName := fmt.Sprintf("%s.Validate%s", structName, method)
	wantErr := len(expectedErrs) != 0
	// Run the validation function, and make sure we expected an error (or not)
	err := validateFn()
	if (err != nil) != wantErr {
		t.Errorf("%s() error = %v, wantErr %v", funcName, err, wantErr)
	}
	// Make sure the error embeds the following expected errors
	validation.TestExpectErrors(t, funcName, err, expectedErrs...)
}

func TestDeployKey_Validate(t *testing.T) {
	tests := []struct {
		name         string
		key          DeployKey
		methods      []validateMethod
		expectedErrs []error
	}{
		{
			name: "valid create",
			key: DeployKey{
				Name: "foo-deploykey",
				Key:  []byte("some-data"),
			},
			methods: []validateMethod{validateCreate},
		},
		{
			name: "valid delete",
			key: DeployKey{
				Name: "foo-deploykey",
				Key:  []byte("some-data"),
			},
			methods: []validateMethod{validateDelete},
		},
		{
			name: "valid create, with all checked fields populated",
			key: DeployKey{
				Name:       "foo-deploykey",
				Key:        []byte("some-data"),
				Repository: newOrgRepoInfoPtr("github.com", "foo-org", nil, "foo-repo"),
			},
			methods: []validateMethod{validateCreate},
		},
		{
			name: "valid delete, with all checked fields populated",
			key: DeployKey{
				Name:       "foo-deploykey",
				Repository: newOrgRepoInfoPtr("github.com", "foo-org", nil, "foo-repo"),
			},
			methods: []validateMethod{validateDelete},
		},
		{
			name: "invalid create, missing name",
			key: DeployKey{
				Key: []byte("some-data"),
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateCreate},
		},
		{
			name:         "invalid delete, missing name",
			key:          DeployKey{},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateDelete},
		},
		{
			name: "invalid create, missing key",
			key: DeployKey{
				Name: "foo-deploykey",
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateCreate},
		},
		{
			name: "invalid create, invalid org info",
			key: DeployKey{
				Name:       "foo-deploykey",
				Key:        []byte("some-data"),
				Repository: newOrgRepoInfoPtr("github.com", "", nil, "foo-repo"),
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateCreate},
		},
		{
			name: "invalid delete, invalid org info",
			key: DeployKey{
				Name:       "foo-deploykey",
				Repository: newOrgRepoInfoPtr("github.com", "", nil, "foo-repo"),
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateDelete},
		},
		{
			name: "invalid create, invalid user repo info",
			key: DeployKey{
				Name:       "foo-deploykey",
				Key:        []byte("some-data"),
				Repository: newUserRepoInfoPtr("github.com", "foo-org", ""),
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateCreate},
		},
		{
			name: "invalid delete, invalid user repo info",
			key: DeployKey{
				Name:       "foo-deploykey",
				Repository: newUserRepoInfoPtr("", "foo-org", "my-repo"),
			},
			expectedErrs: []error{validation.ErrFieldRequired},
			methods:      []validateMethod{validateDelete},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, method := range tt.methods {
				var validateFn validateFunc
				switch method {
				case validateCreate:
					validateFn = tt.key.ValidateCreate
				case validateDelete:
					validateFn = tt.key.ValidateDelete
				default:
					t.Errorf("unknown validate method: %s", method)
					return
				}

				assertValidation(t, "DeployKey", method, validateFn, tt.expectedErrs)
			}
		})
	}
}

func TestRepository_Validate(t *testing.T) {
	unknownRepositoryVisibility := RepositoryVisibility("unknown")
	tests := []struct {
		name         string
		repo         Repository
		methods      []validateMethod
		expectedErrs []error
	}{
		{
			name: "valid org create and update, without enums",
			repo: Repository{
				RepositoryInfo: newOrgRepoInfo("github.com", "foo-org", nil, "foo-repo"),
			},
			methods: []validateMethod{validateCreate, validateUpdate},
		},
		{
			name: "valid user create and update, without enums",
			repo: Repository{
				RepositoryInfo: newUserRepoInfo("github.com", "user", "foo-repo"),
			},
			methods: []validateMethod{validateCreate, validateUpdate},
		},
		{
			name: "valid create and update, with valid enum and description",
			repo: Repository{
				RepositoryInfo: newOrgRepoInfo("github.com", "foo-org", nil, "foo-repo"),
				Description:    StringVar("foo-description"),
				Visibility:     RepositoryVisibilityVar(RepositoryVisibilityPublic),
			},
			methods: []validateMethod{validateCreate, validateUpdate},
		},
		{
			name: "invalid create and update, invalid enum",
			repo: Repository{
				RepositoryInfo: newUserRepoInfo("github.com", "foo-org", "foo-repo"),
				Visibility:     &unknownRepositoryVisibility,
			},
			methods:      []validateMethod{validateCreate, validateUpdate},
			expectedErrs: []error{validation.ErrFieldEnumInvalid},
		},
		{
			name: "invalid create and update, invalid repo info",
			repo: Repository{
				RepositoryInfo: newOrgRepoInfo("github.com", "foo-org", nil, ""),
				Visibility:     RepositoryVisibilityVar(RepositoryVisibilityPrivate),
			},
			methods:      []validateMethod{validateCreate, validateUpdate},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
		{
			name: "invalid create and update, invalid org info",
			repo: Repository{
				RepositoryInfo: newOrgRepoInfo("github.com", "", nil, "foo-repo"), // invalid org name
				Description:    StringVar(""),                                     // description isn't validated, doesn't need any for now
			},
			methods:      []validateMethod{validateCreate, validateUpdate},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, method := range tt.methods {
				var validateFn validateFunc
				switch method {
				case validateCreate:
					validateFn = tt.repo.ValidateCreate
				case validateUpdate:
					validateFn = tt.repo.ValidateUpdate
				default:
					t.Errorf("unknown validate method: %s", method)
					return
				}

				assertValidation(t, "Repository", method, validateFn, tt.expectedErrs)
			}
		})
	}
}

func TestTeamAccess_Validate(t *testing.T) {
	invalidPermission := RepositoryPermission("unknown")
	tests := []struct {
		name         string
		ta           TeamAccess
		methods      []validateMethod
		expectedErrs []error
	}{
		{
			name: "valid create and delete, required field set",
			ta: TeamAccess{
				Name: "foo-team",
			},
			methods: []validateMethod{validateCreate, validateDelete},
		},
		{
			name:         "invalid create and delete, required name",
			ta:           TeamAccess{},
			methods:      []validateMethod{validateCreate, validateDelete},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
		{
			name: "valid create and delete, also including valid repoinfo",
			ta: TeamAccess{
				Name:       "foo-team",
				Repository: newOrgRepoInfoPtr("github.com", "foo-org", nil, "foo-repo"),
			},
			methods: []validateMethod{validateCreate, validateDelete},
		},
		{
			name: "invalid create and delete, invalid repoinfo",
			ta: TeamAccess{
				Name:       "foo-team",
				Repository: newOrgRepoInfoPtr("github.com", "foo-org", nil, ""),
			},
			methods:      []validateMethod{validateCreate, validateDelete},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
		{
			name: "invalid create and delete, invalid orginfo",
			ta: TeamAccess{
				Name:       "foo-team",
				Repository: newOrgRepoInfoPtr("", "foo-org", nil, "foo-repo"),
			},
			methods:      []validateMethod{validateCreate, validateDelete},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
		{
			name: "invalid create and delete, invalid userinfo",
			ta: TeamAccess{
				Name:       "foo-team",
				Repository: newUserRepoInfoPtr("github.com", "", "foo-repo"),
			},
			methods:      []validateMethod{validateCreate, validateDelete},
			expectedErrs: []error{validation.ErrFieldRequired},
		},
		{
			name: "valid create, with valid enum",
			ta: TeamAccess{
				Name:       "foo-team",
				Permission: RepositoryPermissionVar(RepositoryPermissionPull),
			},
			methods: []validateMethod{validateCreate},
		},
		{
			name: "invalid create, invalid enum",
			ta: TeamAccess{
				Name:       "foo-team",
				Permission: &invalidPermission,
			},
			methods:      []validateMethod{validateCreate},
			expectedErrs: []error{validation.ErrFieldEnumInvalid},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, method := range tt.methods {
				var validateFn validateFunc
				switch method {
				case validateCreate:
					validateFn = tt.ta.ValidateCreate
				case validateDelete:
					validateFn = tt.ta.ValidateDelete
				default:
					t.Errorf("unknown validate method: %s", method)
					return
				}

				assertValidation(t, "TeamAccess", method, validateFn, tt.expectedErrs)
			}
		})
	}
}
