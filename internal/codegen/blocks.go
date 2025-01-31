// Copyright (C) 2024 stencil contributors
// Copyright (C) 2022-2023 Outreach Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Description: This file handles parsing of files

package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

// endStatement is a constant for the end of a statement
const endStatement = "EndBlock"

// blockInfo contains information about a block's contents/file location
type blockInfo struct {
	Name, Contents     string
	StartLine, EndLine int
}

// blockPattern is the regex used for parsing block commands.
// For unit testing of this regex and explanation, see https://regex101.com/r/nFgOz0/1
var blockPattern = regexp.MustCompile(`^\s*(///|###|<!---)\s*([a-zA-Z ]+)\(([a-zA-Z0-9 -]+)\)`)

// v2BlockPattern is the new regex for parsing blocks
// For unit testing of this regex and explanation, see https://regex101.com/r/EHkH5O/1
var v2BlockPattern = regexp.MustCompile(`^\s*(//|##|--|<!--)\s{0,1}<<(/?)Stencil::([a-zA-Z ]+)(\([a-zA-Z0-9 -]+\))?>>`)

// parseBlocks reads the blocks from an existing file, potentially adopting blocks based on the source template,
// if so specified
func parseBlocks(filePath string, sourceTemplate *Template) (map[string]*blockInfo, error) {
	f, err := os.Open(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]*blockInfo), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to read blocks from file %q", filePath)
	}
	defer f.Close()

	return parseBlocksInner(f, filePath, sourceTemplate)
}

// parseBlocksInner is the inner implementation of parseBlocks, reusable from inside adoptBlocks to parse blocks
// from the source template contents
// nolint:funlen // Why: Will refactor in the future.
func parseBlocksInner(r io.ReadSeeker, filePath string, sourceTemplate *Template) (map[string]*blockInfo, error) {
	blocks := make(map[string]*blockInfo)
	var curBlock *blockInfo
	scanner := bufio.NewScanner(r)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		matches := blockPattern.FindStringSubmatch(line)
		if len(matches) == 0 {
			// 0: full match
			// 1: comment prefix
			// 2: / if end of block
			// 3: block name
			// 4: block args, if present
			v2Matches := v2BlockPattern.FindStringSubmatch(line)
			if len(v2Matches) == 5 {
				cmd := v2Matches[3]
				if v2Matches[2] == "/" {
					if curBlock == nil {
						return nil, fmt.Errorf("line %d: found closing <</Stencil::Block>> without an opening <<Stencil::Block>>", i+1)
					}

					if cmd == endStatement {
						return nil, fmt.Errorf("line %d: Stencil::EndBlock with a <</, should use <</Stencil::Block>> instead", i+1)
					}

					// If there is a /, it's a closing tag and we should
					// translate it to a closing block command
					cmd = endStatement
					if v2Matches[4] != "" {
						return nil, fmt.Errorf("line %d: expected no arguments to <</Stencil::Block>>", i+1)
					}

					v2Matches[4] = fmt.Sprintf("(%s)", curBlock.Name)
				} else if cmd == endStatement {
					// If it's not a closing tag, but the command is EndBlock,
					// we should error. This is because we don't want to
					// allow users to use the old EndBlock command
					// without a closing tag
					return nil, errors.Errorf("line %d: <<Stencil::EndBlock>> should be <</Stencil::Block>>", i+1)
				}

				// fake the old matches format so we can reuse the same code
				matches = []string{
					v2Matches[0],
					v2Matches[1],
					cmd,
					strings.TrimPrefix(strings.TrimSuffix(v2Matches[4], ")"), "("),
				}
			}
		}
		isCommand := false

		// 1: Comment (###|///)
		// 2: Command
		// 3: Argument to the command
		if len(matches) == 4 {
			cmd := matches[2]
			isCommand = true

			switch cmd {
			case "Block":
				blockName := matches[3]
				if curBlock != nil {
					return nil, fmt.Errorf("invalid Block when already inside of a block, at %s:%d", filePath, i+1)
				}
				curBlock = &blockInfo{
					Name:      blockName,
					StartLine: i,
				}
				blocks[blockName] = curBlock
			case endStatement:
				blockName := matches[3]

				if curBlock == nil {
					return nil, fmt.Errorf("invalid EndBlock when not inside of a block, at %s:%d", filePath, i+1)
				}

				if blockName != curBlock.Name {
					return nil, fmt.Errorf(
						"invalid EndBlock, found EndBlock with name %q while inside of block with name %q, at %s:%d",
						blockName, curBlock.Name, filePath, i+1,
					)
				}

				curBlock.EndLine = i
				curBlock = nil
			default:
				isCommand = false
			}
		}

		// we skip lines that had a recognized command in them, or that
		// aren't in a block
		if isCommand || curBlock == nil {
			continue
		}

		// add the line we processed to the current block we're in
		// and account for having an existing curVal or not. If we
		// don't then we assign curVal to start with the line we
		// just found.
		if curBlock.Contents != "" {
			curBlock.Contents += "\n" + line
		} else {
			curBlock.Contents = line
		}
	}

	if curBlock != nil {
		return nil, fmt.Errorf("found dangling Block (%s) in %s", curBlock.Name, filePath)
	}

	if sourceTemplate != nil && sourceTemplate.adoptMode {
		var err error
		blocks, err = adoptBlocks(r, blocks, sourceTemplate)
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
}

// adoptBlocks adopts the blocks from the source template into the existing blocks
func adoptBlocks(r io.ReadSeeker, blocks map[string]*blockInfo, sourceTemplate *Template) (map[string]*blockInfo, error) {
	tr := bytes.NewReader(sourceTemplate.Contents)
	templateBlocks, err := parseBlocksInner(tr, sourceTemplate.Path, nil)
	if err != nil {
		return nil, err
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// In theory, you can safely split on \n on all OSes because the source line will also have an extra \r on it and match
	templateLines := strings.Split(string(sourceTemplate.Contents), "\n")
	fileLines := strings.Split(string(contents), "\n")

	for k, v := range templateBlocks {
		// Only look for new blocks from the template that aren't already in the file
		if _, ok := blocks[k]; ok {
			continue
		}

		for numLines := 1; ; numLines++ {
			if v.StartLine-numLines < 0 || v.EndLine+numLines >= len(templateLines) {
				break
			}
			preLines := templateLines[v.StartLine-numLines : v.StartLine]
			postLines := templateLines[v.EndLine+1 : v.EndLine+numLines+1]

			// Make sure none of the pre or post lines are block start/ends
			if slices.IndexFunc(preLines, func(x string) bool {
				return blockPattern.MatchString(x) || v2BlockPattern.MatchString(x)
			}) != -1 || slices.IndexFunc(postLines, func(x string) bool {
				return blockPattern.MatchString(x) || v2BlockPattern.MatchString(x)
			}) != -1 {
				continue
			}

			// Search for any potentially-matching pres/posts in the file for this block from the template
			prePositions := findSubsetPositions(fileLines, preLines)
			postPositions := findSubsetPositions(fileLines, postLines)

			if len(prePositions) == 0 || len(postPositions) == 0 {
				break
			}
			// Filter out any posts before the first pre
			_ = slices.DeleteFunc(postPositions, func(x int) bool {
				return x < prePositions[0]+numLines
			})
			if len(postPositions) == 0 {
				break
			}
			// Filter out any pres before the last post
			_ = slices.DeleteFunc(prePositions, func(x int) bool {
				return x > postPositions[len(postPositions)-1]-numLines
			})
			if len(prePositions) == 0 {
				break
			}

			if len(prePositions) > 1 && len(postPositions) > 1 {
				// Heuristic: keep searching until we find exactly one pre or post for the block -- multiple of each
				// means we need to add more lines to the check stack to narrow it down
				continue
			}

			// If there's a single pre left, use that and pick the closest post
			if len(prePositions) == 1 {
				pre := prePositions[0] + numLines - 1
				blocks[k] = &blockInfo{
					Name:      k,
					StartLine: pre,
					EndLine:   postPositions[0],
					Contents:  strings.Join(fileLines[pre+1:postPositions[0]], "\n"),
				}
				break
			}

			// Must be a single post, do the same but in reverse -- use the last preposition with it
			pre := prePositions[len(prePositions)-1] + numLines - 1
			blocks[k] = &blockInfo{
				Name:      k,
				StartLine: pre,
				EndLine:   postPositions[0],
				Contents:  strings.Join(fileLines[pre+1:postPositions[0]], "\n"),
			}
			break
		}
	}

	return blocks, nil
}

func findSubsetPositions(haystack, needles []string) []int {
	res := []int{}
	for i := 0; i < len(haystack)-len(needles)+1; i++ {
		found := true
		for h := 0; h < len(needles); h++ {
			if haystack[i+h] != needles[h] {
				found = false
				break
			}
		}
		if found {
			res = append(res, i)
		}
	}
	return res
}
