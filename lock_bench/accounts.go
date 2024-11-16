package lock_bench

import (
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
)

type Counter struct {
	Count uint64
}

var CounterDiscriminator = [8]byte{255, 176, 4, 245, 188, 253, 124, 25}

func (obj Counter) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Write account discriminator:
	err = encoder.WriteBytes(CounterDiscriminator[:], false)
	if err != nil {
		return err
	}
	// Serialize `Count` param:
	err = encoder.Encode(obj.Count)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Counter) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Read and check account discriminator:
	{
		discriminator, err := decoder.ReadTypeID()
		if err != nil {
			return err
		}
		if !discriminator.Equal(CounterDiscriminator[:]) {
			return fmt.Errorf(
				"wrong discriminator: wanted %s, got %s",
				"[255 176 4 245 188 253 124 25]",
				fmt.Sprint(string(discriminator[:])))
		}
	}
	// Deserialize `Count`:
	err = decoder.Decode(&obj.Count)
	if err != nil {
		return err
	}
	return nil
}
