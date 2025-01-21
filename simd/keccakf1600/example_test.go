package keccakf1600_test

import (
	"encoding/binary"
	"fmt"

	"github.com/cloudflare/circl/internal/sha3"
	"github.com/cloudflare/circl/simd/keccakf1600"
)

func Example() {
	// As an example, computes the (first 32 bytes of a) SHAKE-256 stream of
	// eight short strings at the same time.
	msgs := [8][]byte{
		[]byte("These are some short"),
		[]byte("strings of the same "),
		[]byte("length that fit in a"),
		[]byte("single block.       "),
		[]byte("Another short string."),
		[]byte("Yet another string!  "),
		[]byte("One more for good luck."),
		[]byte("Final string here.   "),
	}
	var hashes [8][32]byte

	// The user could branch to a fast non-SIMD implementation if this function
	// returns false.
	if !keccakf1600.IsEnabledX8() {
		// Compute hashes separately using golang.org/x/crypto/sha3 instead
		// when a fast eight-way implementation is not available. A generic
		// keccakf1600 implementation is quite a bit slower than using
		// the non-interleaved hashes because of the need to interleave and
		// deinterleave the state.
		for i := 0; i < 8; i++ {
			h := sha3.NewShake256()
			_, _ = h.Write(msgs[i])
			_, _ = h.Read(hashes[i][:])
		}
	} else {
		// f1600 acts on 1600 bits arranged as 25 uint64s. Our eight-way f1600
		// acts on eight interleaved states; that is a [200]uint64. (A separate
		// type is used to ensure that the encapsulated [200]uint64 is aligned
		// properly to be used efficiently with vector instructions.)
		var perm keccakf1600.StateX8
		state := perm.Initialize(false)

		// state is initialized with zeroes. As the messages fit within one
		// block, we only need to write the messages, domain separators
		// and padding.
		for i := 0; i < 8; i++ {
			// The messages.
			state[i] = binary.LittleEndian.Uint64(msgs[i][:8])
			state[8+i] = binary.LittleEndian.Uint64(msgs[i][8:16])

			// Final bit of the message together with the SHAKE-256 domain
			// separator (0b1111) and the start of the padding (0b10....)
			state[16+i] = uint64(binary.LittleEndian.Uint32(msgs[i][16:])) |
				(uint64(0x1f) << 32)
			state[32*4+i] = 0x80 << 56 // end of padding (0b...01)
		}

		// Executes the permutation on state.
		perm.Permute()

		// As our desired output fits within one block, we can read it without
		// repeating the permutation.
		for i := 0; i < 8; i++ {
			for j := 0; j < 4; j++ {
				binary.LittleEndian.PutUint64(
					hashes[i][8*j:8*(j+1)],
					state[4*j+i],
				)
			}
		}
	}

	fmt.Printf("\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n%x\n", hashes[0], hashes[1], hashes[2], hashes[3], hashes[4], hashes[5], hashes[6], hashes[7])
	// Output:
	// (Expected output for the hashes)
}
