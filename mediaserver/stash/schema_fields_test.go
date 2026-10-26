package stash

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

// fillAll sets every field of the struct v points to a non-zero value.
func fillAll(t *testing.T, v any) {
	t.Helper()
	setters := map[reflect.Kind]func(f reflect.Value){
		reflect.String:  func(f reflect.Value) { f.SetString("x") },
		reflect.Pointer: func(f reflect.Value) { f.Set(reflect.New(f.Type().Elem())) },
		reflect.Slice:   func(f reflect.Value) { f.Set(reflect.MakeSlice(f.Type(), 1, 1)) },
		reflect.Map: func(f reflect.Value) {
			f.Set(reflect.MakeMapWithSize(f.Type(), 1))
			f.SetMapIndex(reflect.ValueOf("k"), reflect.ValueOf(any("v")))
		},
	}
	rv := reflect.ValueOf(v).Elem()
	for i := range rv.NumField() {
		set, ok := setters[rv.Field(i).Kind()]
		if !ok {
			t.Fatalf("unhandled kind %s for %s", rv.Field(i).Kind(), rv.Type().Field(i).Name)
		}
		set(rv.Field(i))
	}
}

func jsonKeys(t *testing.T, v any) []string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Field names copied from Stash v0.31.1 (commit 4de2351) graphql/schema/types/performer.graphql:67.
func performerInputFields() []string {
	return []string{
		"name", "disambiguation", "url", "urls", "gender", "birthdate", "ethnicity", "country",
		"eye_color", "height_cm", "measurements", "fake_tits", "penis_length", "circumcised",
		"career_length", "career_start", "career_end", "tattoos", "piercings", "alias_list",
		"twitter", "instagram", "favorite", "tag_ids", "image", "stash_ids", "rating100",
		"details", "death_date", "hair_color", "weight", "ignore_auto_tag", "custom_fields",
	}
}

// Stash v0.31.1 graphql/schema/types/performer.graphql:67, performer.graphql:107,
// scene.graphql:129, scene.graphql:88, scene.graphql:93, stash-box.graphql:22, metadata.graphql:369.
func TestMutationInputs_SchemaFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input any
		want  []string
	}{
		{"PerformerCreateInput", &PerformerCreateInput{}, performerInputFields()},
		{"PerformerUpdateInput", &PerformerUpdateInput{}, append([]string{"id"}, performerInputFields()...)},
		{"SceneUpdateInput", &SceneUpdateInput{}, []string{
			"clientMutationId", "id", "title", "code", "details", "director", "url", "urls",
			"date", "rating100", "o_counter", "organized", "studio_id", "gallery_ids",
			"performer_ids", "groups", "movies", "tag_ids", "cover_image", "stash_ids",
			"resume_time", "play_duration", "play_count", "primary_file_id", "custom_fields",
		}},
		{"StashIDInput", &StashIDInput{}, []string{"endpoint", "stash_id", "updated_at"}},
		{"CustomFieldsInput", &CustomFieldsInput{}, []string{"full", "partial", "remove"}},
		{"SceneGroupInput", &SceneGroupInput{}, []string{"group_id", "scene_index"}},
		{"SceneMovieInput", &SceneMovieInput{}, []string{"movie_id", "scene_index"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fillAll(t, tc.input)
			want := slices.Sorted(slices.Values(tc.want))
			if got := jsonKeys(t, tc.input); !slices.Equal(got, want) {
				t.Errorf("json keys = %v\nwant schema fields %v", got, want)
			}
		})
	}
}
