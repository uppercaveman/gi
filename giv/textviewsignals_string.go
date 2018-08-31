// Code generated by "stringer -type=TextViewSignals"; DO NOT EDIT.

package giv

import (
	"fmt"
	"strconv"
)

const _TextViewSignals_name = "TextViewDoneTextViewSelectedTextViewSignalsN"

var _TextViewSignals_index = [...]uint8{0, 12, 28, 44}

func (i TextViewSignals) String() string {
	if i < 0 || i >= TextViewSignals(len(_TextViewSignals_index)-1) {
		return "TextViewSignals(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TextViewSignals_name[_TextViewSignals_index[i]:_TextViewSignals_index[i+1]]
}

func (i *TextViewSignals) FromString(s string) error {
	for j := 0; j < len(_TextViewSignals_index)-1; j++ {
		if s == _TextViewSignals_name[_TextViewSignals_index[j]:_TextViewSignals_index[j+1]] {
			*i = TextViewSignals(j)
			return nil
		}
	}
	return fmt.Errorf("String %v is not a valid option for type TextViewSignals", s)
}