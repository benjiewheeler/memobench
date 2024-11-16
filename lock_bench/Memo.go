package lock_bench

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Memo is the `memo` instruction.
type Memo struct {
	Input *[]byte

	ag_solanago.AccountMetaSlice `bin:"-"`
}

type MemoAccounts struct{}

// NewMemoInstructionBuilder creates a new `Memo` instruction builder.
func NewMemoInstructionBuilder() *Memo {
	nd := &Memo{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 0),
	}
	return nd
}

func (inst *Memo) GetMemoAccounts() *MemoAccounts {
	res := &MemoAccounts{}
	return res
}

// SetInput sets the "input" parameter.
func (inst *Memo) SetInput(input []byte) *Memo {
	inst.Input = &input
	return inst
}

func (inst Memo) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_Memo,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Memo) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Memo) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Input == nil {
			return errors.New("input parameter is not set")
		}
	}

	return nil
}

func (inst *Memo) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Memo")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Input", *inst.Input))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=0]").ParentFunc(func(accountsBranch ag_treeout.Branches) {})
				})
		})
}

func (obj Memo) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Input` param:
	err = encoder.Encode(obj.Input)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Memo) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Input`:
	err = decoder.Decode(&obj.Input)
	if err != nil {
		return err
	}
	return nil
}

// NewMemoInstruction declares a new Memo instruction with the provided parameters and accounts.
func NewMemoInstruction(
	// Parameters:
	input []byte,
) *Memo {
	return NewMemoInstructionBuilder().
		SetInput(input)
}
