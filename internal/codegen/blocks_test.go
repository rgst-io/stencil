package codegen

import (
	"context"
	"os"
	"testing"
	"time"

	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/internal/testing/testmemfs"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

func fakeBlocksTemplate(t *testing.T) *Template {
	return &Template{log: slogext.NewTestLogger(t)}
}

func TestParseBlocks(t *testing.T) {
	blocks, err := parseBlocks("testdata/blocks-test.txt", fakeBlocksTemplate(t))
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	assert.Equal(t, blocks["hello-world"].Contents, "Hello, world!", "expected parseBlocks() to parse basic block")
	assert.Equal(t, blocks["e2e"].Contents, "content", "expected parseBlocks() to parse e2e block")
}

func TestDanglingBlock(t *testing.T) {
	_, err := parseBlocks("testdata/danglingblock-test.txt", fakeBlocksTemplate(t))
	assert.Error(t, err, "found dangling Block (dangles) in testdata/danglingblock-test.txt", "expected parseBlocks() to fail")
}

func TestDanglingEndBlock(t *testing.T) {
	_, err := parseBlocks("testdata/danglingendblock-test.txt", fakeBlocksTemplate(t))
	assert.Error(t, err,
		"invalid EndBlock when not inside of a block, at testdata/danglingendblock-test.txt:8",
		"expected parseBlocks() to fail")
}

func TestBlockInsideBlock(t *testing.T) {
	_, err := parseBlocks("testdata/blockinsideblock-test.txt", fakeBlocksTemplate(t))
	assert.Error(t, err,
		"invalid Block when already inside of a block, at testdata/blockinsideblock-test.txt:3",
		"expected parseBlocks() to fail")
}

func TestWrongEndBlock(t *testing.T) {
	_, err := parseBlocks("testdata/wrongendblock-test.txt", fakeBlocksTemplate(t))
	assert.Error(t, err,
		"invalid EndBlock, found EndBlock with name \"wrongend\" while inside of block with name \"helloWorld\", at testdata/wrongendblock-test.txt:3", //nolint:lll // Why: test
		"expected parseBlocks() to fail")
}

func TestParseV2Blocks(t *testing.T) {
	blocks, err := parseBlocks("testdata/v2blocks-test.txt", fakeBlocksTemplate(t))
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	assert.Equal(t, blocks["helloWorld"].Contents, "Hello, world!", "expected parseBlocks() to parse basic block")
}

func TestV2BlocksErrors(t *testing.T) {
	_, err := parseBlocks("testdata/v2blocks-invalid.txt", fakeBlocksTemplate(t))
	if err == nil {
		t.Fatal("expected parseBlocks() to fail")
	}
}

func TestBasicAdopt(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt1.tpl", "testdata/adopt/adopt1.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 1,
		EndLine:   3,
		Contents:  "  version: xyz",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func TestAdoptWithMultiplePres(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt2.tpl", "testdata/adopt/adopt2.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 1,
		EndLine:   3,
		Contents:  "  version: xyz",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func TestAdoptWithMultiplePresUseNext(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt3.tpl", "testdata/adopt/adopt3.yaml")
	expb := blockInfo{
		Name:      "version",
		StartLine: 5,
		EndLine:   7,
		Contents:  "  version: abc",
	}
	assert.Equal(t, *blocks["version"], expb, "expected parseBlocks() to parse version block")
}

func TestAdoptWithBlockAlreadyPresentDiffName(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adopt4.tpl", "testdata/adopt/adopt4.yaml")
	exp1 := blockInfo{
		Name:      "version1",
		StartLine: 2,
		EndLine:   4,
		Contents:  "  version: xyz",
		Version:   BlockVersion2,
	}
	exp2 := blockInfo{
		Name:      "version2",
		StartLine: 8,
		EndLine:   10,
		Contents:  "  version: abc",
		Version:   BlockVersion2,
	}
	exp := blockInfo{
		Name:      "version",
		StartLine: 7,
		EndLine:   11,
		Contents:  "  ## <<Stencil::Block(version2)>>\n  version: abc\n  ## <</Stencil::Block>>",
		Version:   0, // not present in the file
	}
	assert.DeepEqual(t, *blocks["version1"], exp1)
	assert.DeepEqual(t, *blocks["version2"], exp2)
	assert.DeepEqual(t, *blocks["version"], exp)
}

// This demonstrates that it's not a perfect system, so you might get semi unexpected results if your blocks are too similar
func TestAdoptWithBadBlock(t *testing.T) {
	blocks := adoptTestHelper(t, "testdata/adopt/adoptbad1.tpl", "testdata/adopt/adoptbad1.yaml")
	exp := blockInfo{
		Name:      "version",
		StartLine: 1,
		EndLine:   7,
		Contents:  "  version: xyz\n  otherField: 1\nlocal:\n  deploymentEnvironmentx: prod\n  version: abc",
	}
	assert.Equal(t, *blocks["version"], exp, "expected parseBlocks() to parse wacky version block")
}

func adoptTestHelper(t *testing.T, templateFile, targetFile string) map[string]*blockInfo {
	fs, err := testmemfs.WithManifest("name: testing\n")
	assert.NilError(t, err, "failed to testmemfs.WithManifest")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	log := slogext.NewTestLogger(t)
	assert.NilError(t, err, "failed to NewWithFS")

	conts, err := os.ReadFile(templateFile)
	assert.NilError(t, err, "failed to read templateFile")
	tpl, err := NewTemplate(m, templateFile, 0o644, time.Now(), conts, log, &NewTemplateOpts{
		Adopt: true,
	})
	assert.NilError(t, err, "failed to NewTemplate")

	blocks, err := parseBlocks(targetFile, tpl)
	assert.NilError(t, err, "expected parseBlocks() not to fail")
	return blocks
}
