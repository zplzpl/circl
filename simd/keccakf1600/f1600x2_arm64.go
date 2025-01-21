//go:build arm64 && go1.16 && !purego
// +build arm64,go1.16,!purego

package keccakf1600

import "github.com/cloudflare/circl/internal/sha3"

// Add the new function for permuteSIMDx8
func permuteSIMDx8(state []uint64, turbo bool) { permuteScalarX8(state, turbo) }

func permuteSIMDx2(state []uint64, turbo bool) { f1600x2ARM(&state[0], &sha3.RC, turbo) }

func permuteSIMDx4(state []uint64, turbo bool) { permuteScalarX4(state, turbo) }

//go:noescape
func f1600x2ARM(state *uint64, rc *[24]uint64, turbo bool)
