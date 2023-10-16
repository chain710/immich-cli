package cmd

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"testing"
)

type formStruct struct {
	Int1    *int    `form:"int1,omitempty" json:"jint1,omitempty"`
	String1 *string `form:"string1,omitempty" json:"jstring1,omitempty"`
}

func Test_FlagSetToFormFields(t *testing.T) {
	flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	var s formStruct

	// object's form fields add to flag set
	addFlagSetByFormFields(&s, flagSet)
	// set flag's value
	require.NoError(t, flagSet.Set("int1", "98"))
	require.NoError(t, flagSet.Set("string1", "test1"))
	// flag set value set into object
	require.NoError(t, setFormFields(&s, flagSet))
	require.NotNil(t, s.Int1)
	require.Equal(t, 98, *s.Int1)
	require.Equal(t, "test1", *s.String1)
}
