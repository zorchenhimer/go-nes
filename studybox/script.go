package studybox

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

type Script struct {
	nodes []scriptNode
}

func (s *Script) WriteToFile(filename string) error {
	if len(s.nodes) == 0 {
		return fmt.Errorf("no nodes to output!")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, node := range s.nodes {
		//fmt.Println(node.Asm())
		_, err = fmt.Fprintln(file, node.Asm())
		if err != nil {
			return err
		}
	}
	return nil
}

type scriptNode interface {
	Type() string
	Asm() string
}

type scriptOpCode struct {
	code     uint8
	name     string
	operands []byte
}

func (s *scriptOpCode) Type() string {
	return "OpCode"
}

type scriptOpDefinition struct {
	name string
	argCount int   // inline opcode arguments
	ctrl     uint8 // "parameter control" (to/from stack)
}

func (s *scriptOpCode) Asm() string {
	args := []string{}
	for _, op := range s.operands {
		args = append(args, fmt.Sprintf("$%02X", op))
	}

	argStr := ""
	if len(args) > 0 {
		argStr = " " + strings.Join(args, ", ")
	}

	if s.name != "" {
		return fmt.Sprintf("[%02X] %s%s", s.code, s.name, argStr)
	}
	return fmt.Sprintf("OP_%02X%s", s.code, argStr)
}

type scriptData struct {
	data []byte
}

func (s *scriptData) Type() string {
	return "Data"
}

func (s *scriptData) Asm() string {
	vals := []string{}
	for _, d := range s.data {
		vals = append(vals, fmt.Sprintf("$%02X", d))
	}
	return "data " + strings.Join(vals, ", ")
}

var vmOpCodes = map[byte]scriptOpDefinition{
	// 0x80
	0x80: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x00,
	},

	0x81: scriptOpDefinition{
		name:     "halt_81",
		argCount: 0,
		ctrl:     0x00,
	},

	0x82: scriptOpDefinition{
		name:     "sync_tape_ctrl",
		argCount: 0,
		ctrl:     0x00,
	},

	0x83: scriptOpDefinition{
		name:     "sync_EE",
		argCount: 0,
		ctrl:     0x00,
	},

	0x84: scriptOpDefinition{
		name:     "absolute_jump",
		argCount: 2,
		ctrl:     0x00,
	},

	0x85: scriptOpDefinition{
		name:     "absolute_call",
		argCount: 2,
		ctrl:     0x00,
	},

	0x86: scriptOpDefinition{
		name:     "return",
		argCount: 2,
		ctrl:     0x00,
	},

	0x87: scriptOpDefinition{
		name:     "loop_check?",
		argCount: 0,
		ctrl:     0x00,
	},

	0x88: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x10,
	},

	0x89: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x17,
	},

	0x8A: scriptOpDefinition{
		name:     "read_32",
		argCount: 2,
		ctrl:     0x00,
	},

	0x8B: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0x8C: scriptOpDefinition{
		name:     "str_length",
		argCount: 0,
		ctrl:     0xD0,
	},

	0x8E: scriptOpDefinition{
		name:     "str_concat",
		argCount: 0,
		ctrl:     0x0A,
	},

	0x8F: scriptOpDefinition{
		// comparison?
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	// 0x90
	0x90: scriptOpDefinition{
		// a comparison
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	0x91: scriptOpDefinition{
		// a comparison
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	0x92: scriptOpDefinition{
		// a comparison
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	0x93: scriptOpDefinition{
		// same as 0x92, but negated
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	0x94: scriptOpDefinition{
		// same as 0x91, but negated
		name:     "",
		argCount: 0,
		ctrl:     0xE0,
	},

	0x95: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0x96: scriptOpDefinition{
		// reads inline operand to 0x4E.w
		name:     "set_4E",
		argCount: 2,
		ctrl:     0x00,
	},

	0x97: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x04,
	},

	0x98: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0x99: scriptOpDefinition{
		// disable NMI, screw with the APU
		name:     "",
		argCount: 0,
		ctrl:     0x00,
	},

	0x9A: scriptOpDefinition{
		name:     "disable_apu",
		argCount: 0,
		ctrl:     0x00,
	},

	0x9B: scriptOpDefinition{
		name:     "halt_9B",
		argCount: 0,
		ctrl:     0x00,
	},

	0x9C: scriptOpDefinition{
		// a toggle?
		name:     "toggle_9C",
		argCount: 0,
		ctrl:     0x00,
	},

	0x9D: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x04,
	},

	0x9E: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x04,
	},

	0x9F: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x0D,
	},

	// 0xA0
	0xA0: scriptOpDefinition{
		// a PPU read?
		name:     "",
		argCount: 0,
		ctrl:     0x44,
	},

	0xA1: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xA2: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xA3: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xA4: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xA5: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xA6: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xA7: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x00,
	},

	0xA8: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x0B,
	},

	0xA9: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xAA: scriptOpDefinition{
		name:     "cond_reset_AA",
		argCount: 0,
		ctrl:     0x02,
	},

	0xAB: scriptOpDefinition{
		name:     "cond_reset_AB",
		argCount: 0,
		ctrl:     0x02,
	},

	0xAC: scriptOpDefinition{
		name:     "cond_restore",
		argCount: 0,
		ctrl:     0x02,
	},

	0xAD: scriptOpDefinition{
		name:     "absolute_value",
		argCount: 0,
		ctrl:     0x42,
	},

	0xAE: scriptOpDefinition{
		name:     "check_sign",
		argCount: 0,
		ctrl:     0x42,
	},

	0xAF: scriptOpDefinition{
		name:     "get_first_extra",
		argCount: 0,
		ctrl:     0x42,
	},

	// 0xB0
	0xB0: scriptOpDefinition{
		name:     "create_extra",
		argCount: 0,
		ctrl:     0x82,
	},

	0xB1: scriptOpDefinition{
		// format the arg as an unsigned hex number w/o leading zeros, into a null-terminated string
		name:     "bin_to_hex",
		argCount: 0,
		ctrl:     0x82,
	},

	0xB2: scriptOpDefinition{
		// Reads D2 of 0x4016
		name:     "read_mic",
		argCount: 0,
		ctrl:     0x40,
	},

	0xB3: scriptOpDefinition{
		name:     "",
		argCount: 2, // unconfirmed
		ctrl:     0x0F,
	},

	0xB4: scriptOpDefinition{
		name:     "cond_copy_ptr",
		argCount: 0,
		ctrl:     0x00,
	},

	0xB5: scriptOpDefinition{
		name:     "cond_copy_str_ptr",
		argCount: 0,
		ctrl:     0x00,
	},

	0xB6: scriptOpDefinition{
		name:     "set_cond_copy_ptr",
		argCount: 0,
		ctrl:     0x00,
	},

	0xB7: scriptOpDefinition{
		name:     "push_word_ptr",
		argCount: 2,
		ctrl:     0x00,
	},

	0xB8: scriptOpDefinition{
		name:     "push_word",
		argCount: 2,
		ctrl:     0x00,
	},

	0xB9: scriptOpDefinition{
		// push $xxxx, idx ; where idx is from the stack
		name:     "push_byte_indexed",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBA: scriptOpDefinition{
		name:     "extra_as_result",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBB: scriptOpDefinition{
		name:     "copy_to_stack",
		argCount: -1, // -1 means it's nil terminated
		ctrl:     0x00,
	},

	0xBC: scriptOpDefinition{
		name:     "",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBD: scriptOpDefinition{
		name:     "pop_to_address",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBE: scriptOpDefinition{
		name:     "write_to_offset",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBF: scriptOpDefinition{
		name:     "jump_not_zero",
		argCount: 2,
		ctrl:     0x02,
	},

	// 0xC0
	0xC0: scriptOpDefinition{
		name:     "jump_zero",
		argCount: 2,
		ctrl:     0x02,
	},

	0xC3: scriptOpDefinition{
		name:     "logical_and",
		argCount: 2,
		ctrl:     0x02,
	},

	0xC4: scriptOpDefinition{
		name:     "logical_or",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC5: scriptOpDefinition{
		name:     "compare_not_equal",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC6: scriptOpDefinition{
		name:     "compare_equal",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC7: scriptOpDefinition{
		name:     "compare_greater",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC8: scriptOpDefinition{
		name:     "compare_greater_or_equal",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC9: scriptOpDefinition{
		name:     "compare_less",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCA: scriptOpDefinition{
		name:     "compare_less_or_equal",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCB: scriptOpDefinition{
		name:     "add",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCC: scriptOpDefinition{
		name:     "subtract",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCD: scriptOpDefinition{
		name:     "multiply",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCF: scriptOpDefinition{
		name:     "negate",
		argCount: 0,
		ctrl:     0xC4,
	},

	// 0xD0
	0xD1: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xD4: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xD5: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xDD: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x0B,
	},

	0xDF: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x07,
	},

	// 0xE0
	0xE0: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x44,
	},

	0xE3: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x42,
	},

	0xE5: scriptOpDefinition{
		name:     "clear_some_sprites",
		argCount: 0,
		ctrl:     0x02,
	},

	0xE6: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xE7: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x0F,
	},

	0xE8: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xE9: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x00,
	},

	0xEB: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x09,
	},

	0xEE: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xEF: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x1D,
	},

	// 0xF0
	0xF2: scriptOpDefinition{
		name:     "halt_F2",
		argCount: 0,
		ctrl:     0x00,
	},

	0xF3: scriptOpDefinition{
		name:     "halt_F3",
		argCount: 0,
		ctrl:     0x00,
	},

	0xF4: scriptOpDefinition{
		name:     "halt_F4",
		argCount: 0,
		ctrl:     0x00,
	},

	0xF9: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x40,
	},

	0xFA: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x40,
	},

	0xFB: scriptOpDefinition{
		name:     "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xFE: scriptOpDefinition{
		name:     "",
		argCount: 2,
		ctrl:     0x09,
	},

	0xFF: scriptOpDefinition{
		name:     "",
		argCount: 2,
		ctrl:     0xFF,
	},
}

func DissassembleScript(data []byte) (*Script, error) {
	script := &Script{
		nodes: []scriptNode{},
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("No script data to dissassemble!")
	}

	reader := bytes.NewReader(data)
	var val byte
	var err error
	var good bool = true

	val, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}
	offset := 0

	for good {
		if val&0x80 == 0x80 {
			// op code
			code, ok := vmOpCodes[val]
			if !ok {
				return script, fmt.Errorf("unknown OP Code at offset $%04X: $%02X", offset, val)
			}
			op := &scriptOpCode{
				name: code.name,
				code: val,
			}

			if code.argCount > 0 {
				for i := code.argCount; i > 0; i-- {
					val, err = reader.ReadByte()
					if err != nil {
						return nil, err
					}
					offset++
					op.operands = append(op.operands, val)
				}
			} else if code.argCount == -1 {
				b := 0
				val, err = reader.ReadByte()
				for err == nil && val != 0 {
					b++
					op.operands = append(op.operands, val)
					val, err = reader.ReadByte()
				}
				if err != nil {
					return nil, err
				}
			}

			script.nodes = append(script.nodes, op)
		} else {
			// data
			if len(script.nodes) == 0 {
				script.nodes = append(script.nodes, &scriptData{data: []byte{val}})
			} else if dataNode, ok := script.nodes[len(script.nodes)-1].(*scriptData); ok {
				dataNode.data = append(dataNode.data, val)
			} else {
				script.nodes = append(script.nodes, &scriptData{data: []byte{val}})
			}
		}

		val, err = reader.ReadByte()
		if err != nil {
			good = false
		}
		offset++
	}

	return script, nil
}
