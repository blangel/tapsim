package output

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/txscript"
)

func StackToString(stack [][]byte) []string {
	var str []string
	for i := len(stack) - 1; i >= 0; i-- {
		b := stack[i]

		// The empty element is treated as 0.
		if len(b) == 0 {
			str = append(str, "<>")
			continue
		}
		s := hex.EncodeToString(b)
		str = append(str, s)
	}

	return str
}

func ScriptToString(script []byte) []string {
	str, _ := txscript.DisasmString(script)
	return strings.Split(str, " ")
}

func VmScriptToString(vm *txscript.Engine, scriptIdx int) []string {
	str, _ := vm.DisasmScript(scriptIdx)

	var ss []string
	// Trim prefix.
	l := len("01:0000: ")
	for _, s := range strings.Split(str, "\n") {
		if s == "" {
			continue
		}

		ss = append(ss, s[l:])
	}

	return ss
}

func WitnessToString(witness [][]byte) []string {
	// TODO: reverse order?
	var str []string
	for _, b := range witness {
		s, _ := txscript.DisasmString(b)
		str = append(str, s)
	}

	return str
}

const columnWidth = 40

func ExecutionTable(pc int, script, stack, altStack, witness []string,
	tags map[string]string) string {

	fullWidth := 4 * (columnWidth + 2)
	s := strings.Repeat("-", fullWidth)
	s += "\n"
	s += fmt.Sprintf(" %s| %s| %s| %s\n",
		FixedWidth(columnWidth, "script", tags),
		FixedWidth(columnWidth, "stack", tags),
		FixedWidth(columnWidth, "alt stack", tags),
		FixedWidth(columnWidth, "witness", tags),
	)
	s += strings.Repeat("-", fullWidth)
	s += "\n"

	row := 0
	for {
		scr := ""
		if row < len(script) {
			scr = script[row]
		}

		stk := ""
		if row < len(stack) {
			stk = stack[row]
		}

		alt := ""
		if row < len(altStack) {
			alt = altStack[row]
		}

		wit := ""
		if row < len(witness) {
			wit = witness[row]
		}

		pcC := " "
		if pc == row {
			pcC = ">"

		}

		s += fmt.Sprintf("%s%s| %s| %s| %s\n",
			pcC,
			FixedWidth(columnWidth, scr, tags),
			FixedWidth(columnWidth, stk, tags),
			FixedWidth(columnWidth, alt, tags),
			FixedWidth(columnWidth, wit, tags),
		)

		if scr == "" && stk == "" && alt == "" && wit == "" {
			break
		}

		row++
	}

	s += strings.Repeat("-", 4*(columnWidth+2))
	s += "\n"

	return s
}

func FixedWidth(w int, s string, tags map[string]string) string {
	// If there's a tag, we want to show that at the end, always.
	tagSuffix := ""
	if tag, ok := tags[s]; ok {
		tagSuffix = fmt.Sprintf("(%s)", tag)
	}

	fw := ""
	w = w - len(tagSuffix)
	if w <= 0 {
		panic("tag too long")
	}
	for i := 0; i < w; i++ {
		if i < len(s) {
			// For long elements, we want to print the last few
			// characters for visibility.
			if i >= w-4 {
				fw += string(s[len(s)-(w-i)])
				continue
			}

			if i >= w-7 {
				//if i >= w-3 {
				fw += "."
				continue
			}

			fw += string(s[i])
			continue
		}

		fw += " "
	}

	fw += tagSuffix

	return fw
}
