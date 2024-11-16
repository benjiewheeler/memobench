package lock_bench

import (
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// Lock is the `lock` instruction.
type Lock struct {
	Input *[]byte

	// [0] = [WRITE, SIGNER] payer
	//
	// [1] = [WRITE] counter
	//
	// [2] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

type LockAccounts struct {
	Payer         ag_solanago.PublicKey
	Counter       ag_solanago.PublicKey
	SystemProgram ag_solanago.PublicKey
}

// NewLockInstructionBuilder creates a new `Lock` instruction builder.
func NewLockInstructionBuilder() *Lock {
	nd := &Lock{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

func (inst *Lock) GetLockAccounts() *LockAccounts {
	res := &LockAccounts{}
	if inst.AccountMetaSlice[0] != nil {
		res.Payer = inst.AccountMetaSlice[0].PublicKey
	}
	if inst.AccountMetaSlice[1] != nil {
		res.Counter = inst.AccountMetaSlice[1].PublicKey
	}
	if inst.AccountMetaSlice[2] != nil {
		res.SystemProgram = inst.AccountMetaSlice[2].PublicKey
	}
	return res
}

// SetInput sets the "input" parameter.
func (inst *Lock) SetInput(input []byte) *Lock {
	inst.Input = &input
	return inst
}

// SetPayerAccount sets the "payer" account.
func (inst *Lock) SetPayerAccount(payer ag_solanago.PublicKey) *Lock {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(payer).WRITE().SIGNER()
	return inst
}

// GetPayerAccount gets the "payer" account.
func (inst *Lock) GetPayerAccount() *ag_solanago.AccountMeta {
	return inst.Get(0)
}

// SetCounterAccount sets the "counter" account.
func (inst *Lock) SetCounterAccount(counter ag_solanago.PublicKey) *Lock {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(counter).WRITE()
	return inst
}

// GetCounterAccount gets the "counter" account.
func (inst *Lock) GetCounterAccount() *ag_solanago.AccountMeta {
	return inst.Get(1)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *Lock) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *Lock {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *Lock) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.Get(2)
}

func (inst Lock) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_Lock,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst Lock) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *Lock) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Input == nil {
			return errors.New("input parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Payer is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Counter is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *Lock) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("Lock")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {
					// Parameters of the instruction:
					instructionBranch.Child("Params[len=1]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("Input", *inst.Input))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=3]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("        payer", inst.Get(0)))
						accountsBranch.Child(ag_format.Meta("      counter", inst.Get(1)))
						accountsBranch.Child(ag_format.Meta("systemProgram", inst.Get(2)))
					})
				})
		})
}

func (obj Lock) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Input` param:
	err = encoder.Encode(obj.Input)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Lock) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Input`:
	err = decoder.Decode(&obj.Input)
	if err != nil {
		return err
	}
	return nil
}

// NewLockInstruction declares a new Lock instruction with the provided parameters and accounts.
func NewLockInstruction(
	// Parameters:
	input []byte,
	// Accounts:
	payer ag_solanago.PublicKey,
	counter ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey,
) *Lock {
	return NewLockInstructionBuilder().
		SetInput(input).
		SetPayerAccount(payer).
		SetCounterAccount(counter).
		SetSystemProgramAccount(systemProgram)
}
