// Code generated by "stringer -type TargetMatchResult"; DO NOT EDIT.

package search

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Match-0]
	_ = x[KeyNoMatch-1]
	_ = x[KeyExcluded-2]
	_ = x[ValTooShort-3]
	_ = x[ValTooLong-4]
	_ = x[ValNoMatch-5]
	_ = x[ValFilePath-6]
	_ = x[ValVariable-7]
	_ = x[ValEntropy-8]
}

const _TargetMatchResult_name = "MatchKeyNoMatchKeyExcludedValTooShortValTooLongValNoMatchValFilePathValVariableValEntropy"

var _TargetMatchResult_index = [...]uint8{0, 5, 15, 26, 37, 47, 57, 68, 79, 89}

func (i TargetMatchResult) String() string {
	if i < 0 || i >= TargetMatchResult(len(_TargetMatchResult_index)-1) {
		return "TargetMatchResult(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TargetMatchResult_name[_TargetMatchResult_index[i]:_TargetMatchResult_index[i+1]]
}