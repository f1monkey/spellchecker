package spellchecker

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Spellchecker_Save(t *testing.T) {
	m1 := newSampleSpellchecker()

	filePath := path.Join(t.TempDir(), "spellchecker.bin")
	file, err := os.Create(filePath)
	require.NoError(t, err)
	err = m1.Save(file)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	m2, err := Load(file)
	require.NoError(t, err)

	assert.EqualValues(t, m1.dict.id("green"), m2.dict.id("green"))
	assert.EqualValues(t, m1.dict.maxErrors, m2.dict.maxErrors)
}
