package parameters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		configs []Parameters
	}
	type test struct {
		name    string
		args    args
		want    Parameters
		wantErr *error
		f       func(tt test)
	}

	standard := func(tt test) {
		got, err := Merge(tt.args.configs...)
		if tt.wantErr != nil {
			assert.EqualError(t, err, (*tt.wantErr).Error())
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tt.want, got)
	}

	mustFromVars := func(vars []string) Parameters {
		c, e := FromVars(vars)
		if e != nil {
			t.Fatal("invalid test input")
		}
		return c
	}

	tests := []test{
		{
			name: "empty",
			args: args{
				[]Parameters{},
			},
			want: Parameters{},
			f:    standard,
		}, {
			name: "empty merge",
			args: args{
				[]Parameters{
					{},
					mustFromVars([]string{}),
				},
			},
			want: Parameters{},
			f:    standard,
		}, {
			name: "single",
			args: args{
				[]Parameters{
					{"akey": "avalue"},
				},
			},
			want: Parameters{
				"akey": "avalue",
			},
			f: standard,
		}, {
			name: "two",
			args: args{
				[]Parameters{
					{"akey": "avalue"},
					{"another": "entry"},
				},
			},
			want: Parameters{
				"akey":    "avalue",
				"another": "entry",
			},
			f: standard,
		}, {
			name: "merge",
			args: args{
				[]Parameters{
					{"akey": "avalue"},
					{"akey": "overridden"},
				},
			},
			want: Parameters{
				"akey": "overridden",
			},
			f: standard,
		}, {
			name: "merge with vars",
			args: args{
				[]Parameters{
					{"akey": "avalue"},
					{"akey": "overridden"},
					mustFromVars([]string{"akey=overridden2", "anotherkey=somevalue"}),
				},
			},
			want: Parameters{
				"akey":       "overridden2",
				"anotherkey": "somevalue",
			},
			f: standard,
		}, {
			name: "nested merge with vars",
			args: args{
				[]Parameters{
					{
						"a": Parameters{
							"nested": Parameters{
								"key": "avalue",
							},
						}},
					{
						"a": Parameters{
							"nested": Parameters{
								"key": "overridden",
							},
						},
					},
					mustFromVars([]string{"a.nested.key=overridden2", "anotherkey=somevalue"}),
				},
			},
			want: Parameters{
				"a": Parameters{
					"nested": Parameters{
						"key": "overridden2",
					},
				},
				"anotherkey": "somevalue",
			},
			f: standard,
		},
	}

	logrus.SetLevel(logrus.DebugLevel)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("[%d] %s", i, tt.name), func(t *testing.T) { tt.f(tt) })
	}
}

func TestWithVars(t *testing.T) {
	type args struct {
		extraParams []string
	}
	type test struct {
		name    string
		args    args
		want    Parameters
		wantErr *error
		f       func(tt test)
	}

	standard := func(tt test) {
		got, err := FromVars(tt.args.extraParams)
		if tt.wantErr != nil {
			assert.EqualError(t, err, (*tt.wantErr).Error())
		} else {
			assert.NoError(t, err)
		}
		assert.EqualValues(t, tt.want, got)
	}

	tests := []test{
		{
			name:    "empty",
			args:    args{extraParams: []string{}},
			want:    Parameters{},
			wantErr: nil,
			f:       standard,
		},
		{
			name: "empty value",
			args: args{extraParams: []string{"key="}},
			want: Parameters{
				"key": "",
			},
			wantErr: nil,
			f:       standard,
		},
		{
			name: "simple",
			args: args{extraParams: []string{"key=value"}},
			want: Parameters{
				"key": "value",
			},
			wantErr: nil,
			f:       standard,
		},
		{
			name: "two",
			args: args{extraParams: []string{"key=value", "another=pair"}},
			want: Parameters{
				"key":     "value",
				"another": "pair",
			},
			wantErr: nil,
			f:       standard,
		},
		{
			name: "nested",
			args: args{extraParams: []string{"key.nested=value"}},
			want: Parameters{
				"key": Parameters{
					"nested": "value",
				},
			},
			wantErr: nil,
			f:       standard,
		},
	}

	logrus.SetLevel(logrus.DebugLevel)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("[%d] %s", i, tt.name), func(t *testing.T) { tt.f(tt) })
	}

	t.Run("with spaces", func(t *testing.T) {
		vars := []string{
			`first="a value"`,
			"second.nested='yet another value'",
		}
		want := Parameters{
			"first": "a value",
			"second": Parameters{
				"nested": "yet another value",
			},
		}

		got, err := FromVars(vars)
		assert.NoError(t, err)
		assert.EqualValues(t, want, got)
	})
}

func TestAppendNested(t *testing.T) {
	type args struct {
		key        string
		value      interface{}
		parameters *Parameters
	}
	type test struct {
		name    string
		args    args
		want    *Parameters
		wantErr string
		f       func(tt test)
	}

	standard := func(tt test) {
		got, err := appendNested(tt.args.parameters, tt.args.key, tt.args.value)
		if len(tt.wantErr) > 0 {
			assert.EqualError(t, err, tt.wantErr)
		} else {
			assert.NoError(t, err)
		}
		assert.EqualValues(t, tt.want, got)

		_, err = json.Marshal(got)
		assert.NoError(t, err, "Should marshal to JSON")
	}

	tests := []test{
		{
			name: "empty with empty",
			args: args{
				key:        "",
				value:      nil,
				parameters: &Parameters{},
			},
			f:       standard,
			want:    &Parameters{},
			wantErr: "unexpected empty nestedKey",
		}, {
			name: "nil params",
			args: args{
				key:        "",
				value:      nil,
				parameters: nil,
			},
			f:       standard,
			want:    nil,
			wantErr: "unexpected nil parameters",
		}, {
			name: "empty with nil",
			args: args{
				key:        "key",
				value:      nil,
				parameters: &Parameters{},
			},
			f: standard,
			want: &Parameters{
				"key": nil,
			},
		}, {
			name: "empty with nested nil",
			args: args{
				key:        "key.nested",
				value:      nil,
				parameters: &Parameters{},
			},
			f: standard,
			want: &Parameters{
				"key": Parameters{
					"nested": nil,
				},
			},
		}, {
			name: "empty with nested conflict",
			args: args{
				key:        "key.nested",
				value:      nil,
				parameters: &Parameters{"key": "avalue"},
			},
			f:       standard,
			want:    nil,
			wantErr: "key conflict: key 'key' already exists and is not a map, it has type: 'string'",
		}, {
			name: "empty with double nested nil",
			args: args{
				key:        "key.nested.more",
				value:      nil,
				parameters: &Parameters{},
			},
			f: standard,
			want: &Parameters{
				"key": Parameters{
					"nested": Parameters{
						"more": nil,
					},
				},
			},
		},
	}

	logrus.SetLevel(logrus.DebugLevel)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("[%d] %s", i, tt.name), func(t *testing.T) { tt.f(tt) })
	}
}

func MatchesValidPattern(t *testing.T) {
	input := "key=value"
	groups, ok := VarArgRegexp.MatchGroups(input)

	assert.True(t, ok)
	assert.Equal(t, "key", groups["name"])
	assert.Equal(t, "value", groups["value"])
}

func MatchesPatternWithSpaces(t *testing.T) {
	input := "key=value with spaces"
	groups, ok := VarArgRegexp.MatchGroups(input)

	assert.True(t, ok)
	assert.Equal(t, "key", groups["name"])
	assert.Equal(t, "value with spaces", groups["value"])
}

func DoesNotMatchInvalidPattern(t *testing.T) {
	input := "keyvalue"
	_, ok := VarArgRegexp.MatchGroups(input)

	assert.False(t, ok)
}

func DoesNotMatchEmptyString(t *testing.T) {
	input := ""
	_, ok := VarArgRegexp.MatchGroups(input)

	assert.False(t, ok)
}

func MatchesPatternWithValueHavingEqualSign(t *testing.T) {
 input := "key=value=with=equals"
 groups, ok := VarArgRegexp.MatchGroups(input)

 assert.True(t, ok)
 assert.Equal(t, "key", groups["name"])
 assert.Equal(t, "value=with=equals", groups["value"])
}

func TestVarArgRegexp(t *testing.T) {
  t.Run("MatchesValidPattern", MatchesValidPattern)
  t.Run("MatchesPatternWithSpaces", MatchesPatternWithSpaces)
  t.Run("DoesNotMatchInvalidPattern", DoesNotMatchInvalidPattern)
  t.Run("DoesNotMatchEmptyString", DoesNotMatchEmptyString)
  t.Run("MatchesPatternWithValueHavingEqualSign", MatchesPatternWithValueHavingEqualSign)
}