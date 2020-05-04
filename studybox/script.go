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
		fmt.Println(node.Asm())
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
	name     string
	notes    string
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
		return fmt.Sprintf("%s%s", s.name, argStr)
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
	0x83: scriptOpDefinition{
		name:     "sync_EE",
		notes:    "Some sort of synchronization around $ee (whatever that is)?",
		argCount: 0,
		ctrl:     0x00,
	},

	0x84: scriptOpDefinition{
		name:     "absolute_jump",
		notes:    "",
		argCount: 2,
		ctrl:     0x00,
	},

	0x85: scriptOpDefinition{
		name:     "absolute_call",
		notes:    "Pushes the address of the instruction pointer after the inline operand to the stack, then sets the instruction pointer to the inline operand (absolute subroutine call).",
		argCount: 2,
		ctrl:     0x00,
	},

	0x89: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x17,
	},

	0x95: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0x9D: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x04,
	},

	0x9E: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x04,
	},

	0xA1: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xA3: scriptOpDefinition{
		name:     "",
		notes:    "Looks like some sort of scene control.",
		argCount: 0,
		ctrl:     0x02,
	},

	0xAE: scriptOpDefinition{
		name:     "check_sign",
		notes:    "Returns -1 if argument is negative, 0 if argument is zero, 1 if argument is positive.",
		argCount: 0,
		ctrl:     0x42,
	},

	0xB8: scriptOpDefinition{
		name:     "push_word",
		notes:    "Push the inline operand to the stack as a result.",
		argCount: 2,
		ctrl:     0x00,
	},

	0xBB: scriptOpDefinition{
		name:     "copy_to_stack",
		notes:    `Copy the inline operand to the stack as a 16-word ("extra" argument group?) result.`,
		argCount: -1, // -1 means it's nil terminated
		ctrl:     0x00,
	},

	0xBD: scriptOpDefinition{
		name:     "pop_to_address",
		notes:    "Store the parameter given on the stack at the word address given by the inline operand.",
		argCount: 2, // -1 means it's nil terminated
		ctrl:     0x00,
	},

	0xC0: scriptOpDefinition{
		name:     "jump_zero",
		notes:    "If the argument is zero, sets the instruction pointer to the operand (conditional absolute jump).",
		argCount: 2,
		ctrl:     0x02,
	},

	0xC4: scriptOpDefinition{
		name:     "logical_or",
		notes:    "Logical OR: If either parameter is non-zero, the result is 1, otherwise the result is 0.",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xC6: scriptOpDefinition{
		name:     "compare_equal",
		notes:    "Comparison: If both parameters are equal, the result is 1, otherwise the result is 0.",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xCF: scriptOpDefinition{
		name:     "negate",
		notes:    "Unary negate the argument.",
		argCount: 0,
		ctrl:     0xC4,
	},

	0xD4: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xD5: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xDF: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x07,
	},

	0xE3: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x42,
	},

	0xE5: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xE6: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xE7: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x0F,
	},

	0xE8: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x02,
	},

	0xF2: scriptOpDefinition{
		name:     "halt_F2",
		notes:    "Jumps to itself in a tight loop",
		argCount: 0,
		ctrl:     0x00,
	},

	0xF9: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x40,
	},

	0xFA: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 0,
		ctrl:     0x40,
	},

	0xFE: scriptOpDefinition{
		name:     "",
		notes:    "",
		argCount: 2,
		ctrl:     0x09,
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
	//for _, b := range data {
	var val byte
	var err error
	var good bool = true

	val, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}
	offset := 0

	for good {
		//fmt.Printf("script byte $%02X\n", val)
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
					//fmt.Printf("operand: $%02X\n", val)
					op.operands = append(op.operands, val)
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
