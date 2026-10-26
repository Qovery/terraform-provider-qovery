//go:build unit && !integration
// +build unit,!integration

package qovery

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/qovery/qovery-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qovery/terraform-provider-qovery/client"
	"github.com/qovery/terraform-provider-qovery/internal/domain/project"
	"github.com/qovery/terraform-provider-qovery/internal/domain/secret"
	"github.com/qovery/terraform-provider-qovery/internal/domain/variable"
)

// secretEntry builds a secrets entry. An empty id or key reads as null.
func secretEntry(id, key string, value, version, description types.String) Secret {
	optional := func(s string) types.String {
		if s == "" {
			return types.StringNull()
		}
		return types.StringValue(s)
	}
	return Secret{
		Id:             optional(id),
		Key:            optional(key),
		Value:          value,
		ValueWo:        types.StringNull(),
		ValueWoVersion: version,
		Description:    description,
	}
}

var (
	noValue       = types.StringNull()
	noVersion     = types.StringNull()
	noDescription = types.StringNull()
)

// The resources and the data sources share their model struct, so the secrets attribute of every
// schema must decode into it and encode from it, which only fails at runtime.
func TestSecretsSchemaMatchesTheModels(t *testing.T) {
	t.Parallel()

	secrets := SecretList{
		secretEntry("id-1", "PLAIN", types.StringValue("plain"), noVersion, types.StringValue("d")),
		secretEntry("id-2", "WRITE_ONLY", noValue, types.StringValue("1"), noDescription),
	}.toTerraformList()

	testCases := []struct {
		TestName   string
		Resource   resource.Resource
		DataSource datasource.DataSource
		NewModel   func() any
	}{
		{TestName: "application", Resource: applicationResource{}, DataSource: applicationDataSource{}, NewModel: func() any { return &Application{} }},
		{TestName: "container", Resource: containerResource{}, DataSource: containerDataSource{}, NewModel: func() any { return &Container{} }},
		{TestName: "environment", Resource: environmentResource{}, DataSource: environmentDataSource{}, NewModel: func() any { return &Environment{} }},
		{TestName: "helm", Resource: helmResource{}, DataSource: helmDataSource{}, NewModel: func() any { return &Helm{} }},
		{TestName: "job", Resource: jobResource{}, DataSource: jobDataSource{}, NewModel: func() any { return &Job{} }},
		{TestName: "project", Resource: projectResource{}, DataSource: projectDataSource{}, NewModel: func() any { return &Project{} }},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			resourceSchema := schemaFor(t, tc.Resource)
			got := roundTripSecrets(t, tfsdk.State{Schema: resourceSchema}, tc.NewModel(), secrets)
			assert.Len(t, got.Elements(), 2)
			assert.Equal(t, secrets, got, "the resource keeps the order and the version")

			dataSourceResp := &datasource.SchemaResponse{}
			tc.DataSource.Schema(ctx, datasource.SchemaRequest{}, dataSourceResp)
			require.False(t, dataSourceResp.Diagnostics.HasError())
			got = roundTripSecrets(t, tfsdk.State{Schema: dataSourceResp.Schema}, tc.NewModel(), secretsForDataSource(secrets))
			assert.Len(t, got.Elements(), 2)
		})
	}
}

// roundTripSecrets decodes a value of state's schema, every attribute null, into model, sets its
// Secrets field, encodes it back and returns the secrets attribute it holds.
func roundTripSecrets(t *testing.T, state tfsdk.State, model any, secrets types.List) types.List {
	t.Helper()
	ctx := context.Background()

	objType := state.Schema.Type().TerraformType(ctx).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for name, attrType := range objType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	state.Raw = tftypes.NewValue(objType, values)

	require.False(t, state.Get(ctx, model).HasError(), "decoding the schema into the model")
	reflect.ValueOf(model).Elem().FieldByName("Secrets").Set(reflect.ValueOf(secrets))
	diags := state.Set(ctx, model)
	require.False(t, diags.HasError(), "encoding the model: %v", diags)

	var got types.List
	require.False(t, state.GetAttribute(ctx, path.Root("secrets"), &got).HasError())
	return got
}

func TestSecretsValidator(t *testing.T) {
	t.Parallel()

	set := types.StringValue("value")
	unknown := types.StringUnknown()
	entry := func(key, value, valueWo, version types.String) attr.Value {
		return types.ObjectValueMust(secretWithWriteOnlyAttrTypes, map[string]attr.Value{
			"id":               types.StringNull(),
			"key":              key,
			"value":            value,
			"value_wo":         valueWo,
			"value_wo_version": version,
			"description":      types.StringNull(),
		})
	}
	elementType := types.ObjectType{AttrTypes: secretWithWriteOnlyAttrTypes}
	list := func(entries ...attr.Value) types.List { return types.ListValueMust(elementType, entries) }
	keyA, keyB := types.StringValue("A"), types.StringValue("B")

	testCases := []struct {
		TestName       string
		Secrets        types.List
		ExpectedErrors []string
	}{
		{TestName: "success_null_list", Secrets: types.ListNull(elementType)},
		{TestName: "success_unknown_list", Secrets: types.ListUnknown(elementType)},
		{TestName: "success_plain_and_write_only", Secrets: list(entry(keyA, set, noValue, noVersion), entry(keyB, noValue, set, set))},
		{
			TestName:       "error_missing_value",
			Secrets:        list(entry(keyA, noValue, noValue, noVersion)),
			ExpectedErrors: []string{"secrets[0].value"},
		},
		{
			TestName:       "error_both_values",
			Secrets:        list(entry(keyA, set, set, set)),
			ExpectedErrors: []string{"secrets[0].value_wo"},
		},
		{
			TestName:       "error_write_only_without_version",
			Secrets:        list(entry(keyA, set, noValue, noVersion), entry(keyB, noValue, set, noVersion)),
			ExpectedErrors: []string{"secrets[1].value_wo_version"},
		},
		{
			// An empty string is a value: the API receives it.
			TestName:       "error_empty_write_only_without_version",
			Secrets:        list(entry(keyA, noValue, types.StringValue(""), noVersion)),
			ExpectedErrors: []string{"secrets[0].value_wo_version"},
		},
		{
			TestName:       "error_version_with_plain_value",
			Secrets:        list(entry(keyA, set, noValue, set)),
			ExpectedErrors: []string{"secrets[0].value_wo_version"},
		},
		{
			TestName:       "error_version_only",
			Secrets:        list(entry(keyA, noValue, noValue, set)),
			ExpectedErrors: []string{"secrets[0].value", "secrets[0].value_wo_version"},
		},
		{
			TestName:       "error_duplicate_key",
			Secrets:        list(entry(keyA, set, noValue, noVersion), entry(keyB, set, noValue, noVersion), entry(keyA, noValue, set, set)),
			ExpectedErrors: []string{"secrets[2].key"},
		},
		{
			// An ephemeral value is unknown until the ephemeral resource opens.
			TestName: "success_unknown_write_only_without_version",
			Secrets:  list(entry(keyA, noValue, unknown, noVersion)),
		},
		{TestName: "success_unknown_value", Secrets: list(entry(keyA, unknown, noValue, noVersion))},
		{TestName: "success_unknown_keys", Secrets: list(entry(unknown, set, noValue, noVersion), entry(unknown, set, noValue, noVersion))},
		{TestName: "success_unknown_entry", Secrets: list(types.ObjectUnknown(secretWithWriteOnlyAttrTypes), entry(keyA, set, noValue, noVersion))},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			resp := &validator.ListResponse{}
			secretsValidator{}.ValidateList(context.Background(), validator.ListRequest{Path: path.Root("secrets"), ConfigValue: tc.Secrets}, resp)
			expected := tc.ExpectedErrors
			if expected == nil {
				expected = []string{}
			}
			assert.Equal(t, expected, errorPaths(resp.Diagnostics), "diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestSecretsValidator_DuplicateKeyMessage(t *testing.T) {
	t.Parallel()

	duplicate := SecretList{
		secretEntry("", "A", types.StringValue("one"), noVersion, noDescription),
		secretEntry("", "A", types.StringValue("two"), noVersion, noDescription),
	}.toTerraformList()
	resp := &validator.ListResponse{}
	secretsValidator{}.ValidateList(context.Background(), validator.ListRequest{Path: path.Root("secrets"), ConfigValue: duplicate}, resp)

	require.Len(t, resp.Diagnostics.Errors(), 1)
	assert.Equal(t, "secrets[1].key is the key of secrets[0] too: each key can appear once.", resp.Diagnostics.Errors()[0].Detail())
}

// sentSecret is a secret request of either write path: the client one (application) and the
// domain one (the other services).
type sentSecret struct {
	ID    string
	Key   string
	Value string
}

type sentSecrets struct {
	Create []sentSecret
	Update []sentSecret
	Delete []string
}

func clientSecretsSent(diff client.SecretsDiff) sentSecrets {
	sent := sentSecrets{Create: []sentSecret{}, Update: []sentSecret{}, Delete: []string{}}
	for _, c := range diff.Create {
		sent.Create = append(sent.Create, sentSecret{Key: c.Key, Value: derefString(c.Value)})
	}
	for _, u := range diff.Update {
		sent.Update = append(sent.Update, sentSecret{ID: u.Id, Key: u.Key, Value: derefString(u.Value)})
	}
	for _, d := range diff.Delete {
		sent.Delete = append(sent.Delete, d.Id)
	}
	return sent
}

func domainSecretsSent(diff secret.DiffRequest) sentSecrets {
	sent := sentSecrets{Create: []sentSecret{}, Update: []sentSecret{}, Delete: []string{}}
	for _, c := range diff.Create {
		sent.Create = append(sent.Create, sentSecret{Key: c.Key, Value: c.Value})
	}
	for _, u := range diff.Update {
		sent.Update = append(sent.Update, sentSecret{ID: u.SecretID, Key: u.Key, Value: u.Value})
	}
	for _, d := range diff.Delete {
		sent.Delete = append(sent.Delete, d.SecretID)
	}
	return sent
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func TestSecretListDiff_WriteOnly(t *testing.T) {
	t.Parallel()

	v1, v2 := types.StringValue("1"), types.StringValue("2")
	testCases := []struct {
		TestName string
		Old      SecretList
		Planned  SecretList
		ValuesWo writeOnlySecretValues
		Expected sentSecrets
	}{
		{
			TestName: "create_sends_the_write_only_value",
			Planned:  SecretList{secretEntry("", "A", noValue, v1, noDescription)},
			ValuesWo: writeOnlySecretValues{"A": types.StringValue("first")},
			Expected: sentSecrets{Create: []sentSecret{{Key: "A", Value: "first"}}},
		},
		{
			TestName: "version_bump_sends_the_new_value",
			Old:      SecretList{secretEntry("id-a", "A", noValue, v1, noDescription)},
			Planned:  SecretList{secretEntry("", "A", noValue, v2, noDescription)},
			ValuesWo: writeOnlySecretValues{"A": types.StringValue("second")},
			Expected: sentSecrets{Update: []sentSecret{{ID: "id-a", Key: "A", Value: "second"}}},
		},
		{
			TestName: "new_value_alone_sends_nothing",
			Old:      SecretList{secretEntry("id-a", "A", noValue, v1, noDescription)},
			Planned:  SecretList{secretEntry("", "A", noValue, v1, noDescription)},
			ValuesWo: writeOnlySecretValues{"A": types.StringValue("new")},
		},
		{
			// The trigger-only contract: an update of the entry sends the current value.
			TestName: "description_change_resends_the_current_value",
			Old:      SecretList{secretEntry("id-a", "A", noValue, v1, types.StringValue("old"))},
			Planned:  SecretList{secretEntry("", "A", noValue, v1, types.StringValue("new"))},
			ValuesWo: writeOnlySecretValues{"A": types.StringValue("current")},
			Expected: sentSecrets{Update: []sentSecret{{ID: "id-a", Key: "A", Value: "current"}}},
		},
		{
			TestName: "switch_to_write_only_updates_in_place",
			Old:      SecretList{secretEntry("id-a", "A", types.StringValue("plain"), noVersion, noDescription)},
			Planned:  SecretList{secretEntry("", "A", noValue, v1, noDescription)},
			ValuesWo: writeOnlySecretValues{"A": types.StringValue("write-only")},
			Expected: sentSecrets{Update: []sentSecret{{ID: "id-a", Key: "A", Value: "write-only"}}},
		},
		{
			TestName: "switch_to_plain_updates_in_place",
			Old:      SecretList{secretEntry("id-a", "A", noValue, v1, noDescription)},
			Planned:  SecretList{secretEntry("", "A", types.StringValue("plain"), noVersion, noDescription)},
			Expected: sentSecrets{Update: []sentSecret{{ID: "id-a", Key: "A", Value: "plain"}}},
		},
		{
			TestName: "reorder_sends_nothing",
			Old: SecretList{
				secretEntry("id-a", "A", types.StringValue("a"), noVersion, noDescription),
				secretEntry("id-b", "B", noValue, v1, noDescription),
			},
			Planned: SecretList{
				secretEntry("", "B", noValue, v1, noDescription),
				secretEntry("", "A", types.StringValue("a"), noVersion, noDescription),
			},
			ValuesWo: writeOnlySecretValues{"B": types.StringValue("b")},
		},
		{
			TestName: "mixed_entries_delete_and_create",
			Old: SecretList{
				secretEntry("id-a", "A", types.StringValue("a"), noVersion, noDescription),
				secretEntry("id-b", "B", noValue, v1, noDescription),
			},
			Planned: SecretList{
				secretEntry("", "A", types.StringValue("a"), noVersion, noDescription),
				secretEntry("", "C", noValue, v1, noDescription),
				secretEntry("", "D", types.StringValue("d"), noVersion, noDescription),
			},
			ValuesWo: writeOnlySecretValues{"C": types.StringValue("c")},
			Expected: sentSecrets{
				Create: []sentSecret{{Key: "C", Value: "c"}, {Key: "D", Value: "d"}},
				Delete: []string{"id-b"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			expected := tc.Expected
			if expected.Create == nil {
				expected.Create = []sentSecret{}
			}
			if expected.Update == nil {
				expected.Update = []sentSecret{}
			}
			if expected.Delete == nil {
				expected.Delete = []string{}
			}
			assert.Equal(t, expected, clientSecretsSent(tc.Planned.diff(tc.Old, tc.ValuesWo)), "client path")
			assert.Equal(t, expected, domainSecretsSent(tc.Planned.diffRequest(tc.Old, tc.ValuesWo)), "domain path")
		})
	}
}

func TestSecretListConverters_KeepThePriorOrder(t *testing.T) {
	t.Parallel()

	prior := SecretList{
		secretEntry("id-b", "B", types.StringValue("b"), noVersion, noDescription),
		secretEntry("id-a", "A", noValue, types.StringValue("3"), noDescription),
	}.toTerraformList()
	nullPrior := types.ListNull(types.ObjectType{AttrTypes: secretWithWriteOnlyAttrTypes})

	secretType := "VALUE"
	apiSecretType := qovery.APIVARIABLETYPEENUM_VALUE
	clientSecrets := []*qovery.Secret{}
	domainSecrets := secret.Secrets{}
	for i, key := range []string{"A", "B", "C"} {
		clientSecrets = append(clientSecrets, &qovery.Secret{Id: testVariableID(i).String(), Key: key, Scope: qovery.APIVARIABLESCOPEENUM_APPLICATION, VariableType: &apiSecretType})
		domainSecrets = append(domainSecrets, secret.Secret{ID: testVariableID(i), Key: key, Scope: variable.ScopeContainer, Type: secretType})
	}

	converters := map[string]func(secretCollection) SecretList{
		"client": func(prior secretCollection) SecretList {
			return fromSecretList(prior, clientSecrets, qovery.APIVARIABLESCOPEENUM_APPLICATION, secretType)
		},
		"domain": func(prior secretCollection) SecretList {
			return convertDomainSecretsToSecretList(prior, domainSecrets, variable.ScopeContainer, secretType)
		},
	}
	for name, convert := range converters {
		got := convert(prior)
		require.Len(t, got, 3, name)
		assert.Equal(t, []string{"B", "A", "C"}, secretKeys(got), "%s: prior order first, then the new keys", name)
		assert.Equal(t, types.StringValue("b"), got[0].Value, "%s: a plain entry keeps its value", name)
		assert.Equal(t, types.StringNull(), got[1].Value, "%s: a write-only entry has no value", name)
		assert.Equal(t, types.StringValue("3"), got[1].ValueWoVersion, "%s: a write-only entry keeps its version", name)
		assert.Equal(t, types.StringNull(), got[2].ValueWoVersion, "%s: a new key has no version", name)
		for _, s := range got {
			assert.Equal(t, types.StringNull(), s.ValueWo, "%s: the write-only value never reaches the state", name)
		}

		imported := convert(nullPrior)
		assert.Equal(t, []string{"A", "B", "C"}, secretKeys(imported), "%s: import keeps the API order", name)
		for _, s := range imported {
			assert.Equal(t, types.StringNull(), s.Value, "%s: import records no value", name)
			assert.Equal(t, types.StringNull(), s.ValueWoVersion, "%s: import records no version", name)
		}
	}
}

func secretKeys(list SecretList) []string {
	keys := make([]string, 0, len(list))
	for _, s := range list {
		keys = append(keys, ToString(s.Key))
	}
	return keys
}

func TestSecretsForDataSource(t *testing.T) {
	t.Parallel()

	list := SecretList{secretEntry("id-a", "A", noValue, types.StringValue("1"), types.StringValue("d"))}.toTerraformList()
	got := secretsForDataSource(list)

	require.Len(t, got.Elements(), 1)
	attributes := got.Elements()[0].(types.Object).Attributes()
	assert.Equal(t, []string{"description", "id", "key", "value"}, sortedAttributeNames(attributes))
	assert.Equal(t, types.StringValue("A"), attributes["key"])
	assert.True(t, secretsForDataSource(types.ListNull(types.ObjectType{AttrTypes: secretWithWriteOnlyAttrTypes})).IsNull())
}

func sortedAttributeNames(m map[string]attr.Value) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// projectValue builds a value of the project schema, named "project", with the given secrets
// entries, every other attribute null. Each entry maps a secrets attribute to its value.
func projectValue(t *testing.T, id tftypes.Value, entries []map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	objType := schemaFor(t, projectResource{}).Type().TerraformType(context.Background()).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for attrName, attrType := range objType.AttributeTypes {
		values[attrName] = tftypes.NewValue(attrType, nil)
	}
	values["id"] = id
	values["organization_id"] = tftypes.NewValue(tftypes.String, writeOnlyTestOrganizationID)
	values["name"] = tftypes.NewValue(tftypes.String, "project")

	listType := objType.AttributeTypes["secrets"].(tftypes.List)
	elementType := listType.ElementType.(tftypes.Object)
	elements := make([]tftypes.Value, 0, len(entries))
	for _, entry := range entries {
		entryValues := map[string]tftypes.Value{}
		for attrName, attrType := range elementType.AttributeTypes {
			entryValues[attrName] = tftypes.NewValue(attrType, nil)
		}
		for attrName, value := range entry {
			entryValues[attrName] = value
		}
		elements = append(elements, tftypes.NewValue(elementType, entryValues))
	}
	values["secrets"] = tftypes.NewValue(listType, elements)
	return tftypes.NewValue(objType, values)
}

func TestSecretValuesWoFromConfig(t *testing.T) {
	t.Parallel()

	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	config := tfsdk.Config{Schema: schemaFor(t, projectResource{}), Raw: projectValue(t, tftypes.NewValue(tftypes.String, nil), []map[string]tftypes.Value{
		{"key": str("PLAIN"), "value": str("plain")},
		{"key": str("WRITE_ONLY"), "value_wo": str("secret"), "value_wo_version": str("1")},
	})}

	planned := SecretList{
		secretEntry("", "PLAIN", types.StringValue("plain"), noVersion, noDescription),
		secretEntry("", "WRITE_ONLY", noValue, types.StringValue("1"), noDescription),
	}

	var diags diag.Diagnostics
	got := secretValuesWoFromConfig(context.Background(), config, planned, nil, &diags)

	require.False(t, diags.HasError(), "diagnostics: %v", diags)
	assert.Equal(t, writeOnlySecretValues{"WRITE_ONLY": types.StringValue("secret")}, got)
}

// ignore_changes on secrets can plan a key the configuration no longer has. Sending that entry
// would blank the secret, so the apply fails instead, unless the entry is left unchanged.
func TestSecretValuesWoFromConfig_RejectsASecretSentWithoutValue(t *testing.T) {
	t.Parallel()

	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	// The configuration renamed TOKEN to NEW_TOKEN, and ignore_changes kept TOKEN in the plan.
	config := tfsdk.Config{Schema: schemaFor(t, projectResource{}), Raw: projectValue(t, tftypes.NewValue(tftypes.String, nil), []map[string]tftypes.Value{
		{"key": str("NEW_TOKEN"), "value_wo": str("secret"), "value_wo_version": str("2")},
	})}
	v1, v2 := types.StringValue("1"), types.StringValue("2")

	testCases := []struct {
		TestName    string
		Planned     SecretList
		Old         SecretList
		ExpectError bool
	}{
		{
			TestName:    "error_update_without_value",
			Planned:     SecretList{secretEntry("", "TOKEN", noValue, v2, noDescription)},
			Old:         SecretList{secretEntry("id-token", "TOKEN", noValue, v1, noDescription)},
			ExpectError: true,
		},
		{
			TestName:    "error_create_without_value",
			Planned:     SecretList{secretEntry("", "TOKEN", noValue, v2, noDescription)},
			ExpectError: true,
		},
		{
			TestName: "success_unchanged_entry_without_value",
			Planned:  SecretList{secretEntry("", "TOKEN", noValue, v1, noDescription)},
			Old:      SecretList{secretEntry("id-token", "TOKEN", noValue, v1, noDescription)},
		},
		{
			TestName: "success_update_with_the_config_value",
			Planned:  SecretList{secretEntry("", "NEW_TOKEN", noValue, v2, noDescription)},
			Old:      SecretList{secretEntry("id-token", "NEW_TOKEN", noValue, v1, noDescription)},
		},
		{
			TestName: "success_plain_value",
			Planned:  SecretList{secretEntry("", "TOKEN", types.StringValue("plain"), noVersion, noDescription)},
			Old:      SecretList{secretEntry("id-token", "TOKEN", noValue, v1, noDescription)},
		},
		{
			// Keys unknown at plan time skip the duplicate check of the validator.
			TestName: "error_duplicate_key",
			Planned: SecretList{
				secretEntry("", "TOKEN", types.StringValue("one"), noVersion, noDescription),
				secretEntry("", "TOKEN", types.StringValue("two"), noVersion, noDescription),
			},
			ExpectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.TestName, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			secretValuesWoFromConfig(context.Background(), config, tc.Planned, tc.Old, &diags)
			assert.Equal(t, tc.ExpectError, diags.HasError(), "diagnostics: %v", diags)
			if tc.ExpectError {
				assert.Contains(t, diags.Errors()[0].Detail(), `"TOKEN"`)
			}
		})
	}
}

// fakeProjectService applies the secret requests it receives to the secrets it holds, and records
// them.
type fakeProjectService struct {
	project.Service
	requests []project.UpsertServiceRequest
	secrets  secret.Secrets
}

func (f *fakeProjectService) Create(_ context.Context, organizationID string, request project.UpsertServiceRequest) (*project.Project, error) {
	return f.apply(uuid.NewString(), organizationID, request), nil
}

func (f *fakeProjectService) Update(_ context.Context, projectID string, request project.UpsertServiceRequest) (*project.Project, error) {
	return f.apply(projectID, writeOnlyTestOrganizationID, request), nil
}

func (f *fakeProjectService) apply(projectID, organizationID string, request project.UpsertServiceRequest) *project.Project {
	f.requests = append(f.requests, request)
	for _, d := range request.Secrets.Delete {
		kept := secret.Secrets{}
		for _, s := range f.secrets {
			if s.ID.String() != d.SecretID {
				kept = append(kept, s)
			}
		}
		f.secrets = kept
	}
	for _, c := range request.Secrets.Create {
		f.secrets = append(f.secrets, secret.Secret{ID: uuid.New(), Key: c.Key, Description: c.Description, Scope: variable.ScopeProject, Type: "VALUE"})
	}
	return &project.Project{
		ID:             uuid.MustParse(projectID),
		OrganizationID: uuid.MustParse(organizationID),
		Name:           request.ProjectUpsertRequest.Name,
		Secrets:        f.secrets,
	}
}

// The project resource sends secrets[].value_wo from the config, and its state never holds it.
func TestProjectResource_WriteOnlySecret(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	service := &fakeProjectService{}
	r := projectResource{projectService: service}
	s := schemaFor(t, r)
	str := func(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
	null := tftypes.NewValue(tftypes.String, nil)
	unknownID := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)

	// Create: config.value_wo is set, the plan holds null as the framework nulls it.
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}
	r.Create(ctx, resource.CreateRequest{
		Config: tfsdk.Config{Schema: s, Raw: projectValue(t, null, []map[string]tftypes.Value{
			{"key": str("TOKEN"), "value_wo": str("first"), "value_wo_version": str("1")},
		})},
		Plan: tfsdk.Plan{Schema: s, Raw: projectValue(t, unknownID, []map[string]tftypes.Value{
			{"id": unknownID, "key": str("TOKEN"), "value_wo_version": str("1")},
		})},
	}, resp)
	require.False(t, resp.Diagnostics.HasError(), "create: %v", resp.Diagnostics)
	require.Len(t, service.requests[0].Secrets.Create, 1)
	assert.Equal(t, "first", service.requests[0].Secrets.Create[0].Value)

	var created Project
	require.False(t, resp.State.Get(ctx, &created).HasError())
	createdSecrets := ToSecretList(created.Secrets)
	require.Len(t, createdSecrets, 1)
	assert.Equal(t, types.StringNull(), createdSecrets[0].ValueWo)
	assert.Equal(t, types.StringNull(), createdSecrets[0].Value)
	assert.Equal(t, types.StringValue("1"), createdSecrets[0].ValueWoVersion)

	// Update: bumping the version sends the new value.
	projectID := tftypes.NewValue(tftypes.String, created.Id.ValueString())
	updateResp := &resource.UpdateResponse{State: resp.State}
	r.Update(ctx, resource.UpdateRequest{
		Config: tfsdk.Config{Schema: s, Raw: projectValue(t, null, []map[string]tftypes.Value{
			{"key": str("TOKEN"), "value_wo": str("second"), "value_wo_version": str("2")},
		})},
		Plan: tfsdk.Plan{Schema: s, Raw: projectValue(t, projectID, []map[string]tftypes.Value{
			{"id": unknownID, "key": str("TOKEN"), "value_wo_version": str("2")},
		})},
		State: resp.State,
	}, updateResp)
	require.False(t, updateResp.Diagnostics.HasError(), "update: %v", updateResp.Diagnostics)
	require.Len(t, service.requests[1].Secrets.Update, 1)
	assert.Equal(t, "second", service.requests[1].Secrets.Update[0].Value)
	assert.Equal(t, createdSecrets[0].Id.ValueString(), service.requests[1].Secrets.Update[0].SecretID)

	var updated Project
	require.False(t, updateResp.State.Get(ctx, &updated).HasError())
	updatedSecrets := ToSecretList(updated.Secrets)
	require.Len(t, updatedSecrets, 1)
	assert.Equal(t, types.StringNull(), updatedSecrets[0].ValueWo)
	assert.Equal(t, types.StringValue("2"), updatedSecrets[0].ValueWoVersion)
	assert.Equal(t, createdSecrets[0].Id, updatedSecrets[0].Id, "the secret keeps its id")
}
