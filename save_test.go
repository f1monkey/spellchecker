package spellchecker

import (
	"os"
	"path"
	"testing"

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

	require.EqualValues(t, m1.dict.id("green"), m2.dict.id("green"))
	require.EqualValues(t, m1.dict.maxErrors, m2.dict.maxErrors)
	require.EqualValues(t, m1.dict.nextID(), m2.dict.nextID())

	matches := m2.dict.find("orange", 1)
	require.Len(t, matches, 1)
	require.Equal(t, matches[0].Value, "orange")
	require.Greater(t, matches[0].Score, 0.0)
}
