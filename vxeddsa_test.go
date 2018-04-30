package vxeddsa

import (
	"bytes"
	"crypto/sha512"
	"testing"

	"github.com/Scratch-net/vxeddsa/edwards25519"
)

func TestHonestComplete(t *testing.T) {
	sk, err := GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pk := sk.Public()
	alice := []byte("alice")
	aliceVRF := sk.Compute(alice)
	aliceVRFFromProof, aliceProof := sk.Prove(alice)

	// fmt.Printf("pk:           %X\n", pk)
	// fmt.Printf("sk:           %X\n", *sk)
	// fmt.Printf("alice(bytes): %X\n", alice)
	// fmt.Printf("aliceVRF:     %X\n", aliceVRF)
	// fmt.Printf("aliceProof:   %X\n", aliceProof)

	if !pk.Verify(alice, aliceVRF, aliceProof) {
		t.Error("Gen -> Compute -> Prove -> Verify -> FALSE")
	}
	if !bytes.Equal(aliceVRF, aliceVRFFromProof) {
		t.Error("Compute != Prove")
	}
}

func TestFlipBitForgery(t *testing.T) {
	sk, err := GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pk := sk.Public()
	alice := []byte("alice")
	for i := 0; i < 32; i++ {
		for j := uint(0); j < 8; j++ {
			aliceVRF := sk.Compute(alice)
			aliceVRF[i] ^= 1 << j
			_, aliceProof := sk.Prove(alice)
			if pk.Verify(alice, aliceVRF, aliceProof) {
				t.Fatalf("forged by using aliceVRF[%d]^=%d:\n (sk=%x)", i, j, sk)
			}
		}
	}
}

func TestVxed25519Vectors(t *testing.T) {
	signature_correct := [96]byte{
		0x23, 0xc6, 0xe5, 0x93, 0x3f, 0xcd, 0x56, 0x47,
		0x7a, 0x86, 0xc9, 0x9b, 0x76, 0x2c, 0xb5, 0x24,
		0xc3, 0xd6, 0x05, 0x55, 0x38, 0x83, 0x4d, 0x4f,
		0x8d, 0xb8, 0xf0, 0x31, 0x07, 0xec, 0xeb, 0xa0,
		0xa0, 0x01, 0x50, 0xb8, 0x4c, 0xbb, 0x8c, 0xcd,
		0x23, 0xdc, 0x65, 0xfd, 0x0e, 0x81, 0xb2, 0x86,
		0x06, 0xa5, 0x6b, 0x0c, 0x4f, 0x53, 0x6d, 0xc8,
		0x8b, 0x8d, 0xc9, 0x04, 0x6e, 0x4a, 0xeb, 0x08,
		0xce, 0x08, 0x71, 0xfc, 0xc7, 0x00, 0x09, 0xa4,
		0xd6, 0xc0, 0xfd, 0x2d, 0x1a, 0xe5, 0xb6, 0xc0,
		0x7c, 0xc7, 0x22, 0x3b, 0x69, 0x59, 0xa8, 0x26,
		0x2b, 0x57, 0x78, 0xd5, 0x46, 0x0e, 0x0f, 0x05,
	}
	var privKey [32]byte
	var vrfOutPrev, vrfOut [32]byte
	privKey[8] = 189
	edwards25519.ScClamp(&privKey)
	msgLen := 200
	msg := make([]byte, msgLen)
	r := bytes.NewReader(privKey[:])
	sk, _ := GenerateKey(r)

	random := make([]byte, 64)
	signature := sk.signInternal(msg, bytes.NewReader(random))
	if !bytes.Equal(signature, signature_correct[:]) {
		t.Fatal("VXEdDSA sign failed")
	}
	var ok bool
	if ok, vrfOut = sk.Public().verifyInteral(msg, signature); !ok {
		t.Fatal("VXEdDSA verify #1 failed")
	}

	copy(vrfOutPrev[:], vrfOut[:])
	signature[0] ^= 1
	if ok, _ := sk.Public().verifyInteral(msg, signature); ok {
		t.Fatal("VXEdDSA verify #2 should have failed!")
	}
	var sigPrev [96]byte
	copy(sigPrev[:], signature[:])
	sigPrev[0] ^= 1 // undo prev disturbance

	random[0] ^= 1
	signature = sk.signInternal(msg, bytes.NewReader(random))
	if ok, vrfOut = sk.Public().verifyInteral(msg, signature); !ok {
		t.Fatal("VXEdDSA verify #3 failed")
	}
	if !bytes.Equal(vrfOut[:], vrfOutPrev[:]) {
		t.Fatal("VXEdDSA VRF value has changed")
	}
	if bytes.Equal(signature[32:96], sigPrev[32:96]) {
		t.Fatal("VXEdDSA (h, s) changed")
	}
}

func TestVxed25519VectorsSlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped slow test in short mode")
	}
	signature_10k_correct := [96]byte{
		0xa1, 0x96, 0x96, 0xe5, 0x87, 0x3f, 0x6e, 0x5c,
		0x2e, 0xd3, 0x73, 0xab, 0x04, 0x0c, 0x1f, 0x26,
		0x3c, 0xca, 0x52, 0xc4, 0x7e, 0x49, 0xaa, 0xce,
		0xb5, 0xd6, 0xa2, 0x29, 0x46, 0x3f, 0x1b, 0x54,
		0x45, 0x94, 0x9b, 0x6c, 0x27, 0xf9, 0x2a, 0xed,
		0x17, 0xa4, 0x72, 0xbf, 0x35, 0x37, 0xc1, 0x90,
		0xac, 0xb3, 0xfd, 0x2d, 0xf1, 0x01, 0x05, 0xbe,
		0x56, 0x5c, 0xaf, 0x63, 0x65, 0xad, 0x38, 0x04,
		0x70, 0x53, 0xdf, 0x2b, 0xc1, 0x45, 0xc8, 0xee,
		0x02, 0x0d, 0x2b, 0x22, 0x23, 0x7a, 0xbf, 0xfa,
		0x43, 0x31, 0xb3, 0xac, 0x26, 0xd9, 0x76, 0xfc,
		0xfe, 0x30, 0xa1, 0x7c, 0xce, 0x10, 0x67, 0x0e,
	}

	signature_100k_correct := [96]byte{
		0xc9, 0x11, 0x2b, 0x55, 0xfa, 0xc4, 0xb2, 0xfe,
		0x00, 0x7d, 0xf6, 0x45, 0xcb, 0xd2, 0x73, 0xc9,
		0x43, 0xba, 0x20, 0xf6, 0x9c, 0x18, 0x84, 0xef,
		0x6c, 0x65, 0x7a, 0xdb, 0x49, 0xfc, 0x1e, 0xbe,
		0x31, 0xb3, 0xe6, 0xa4, 0x68, 0x2f, 0xd0, 0x30,
		0x81, 0xfc, 0x0d, 0xcd, 0x2d, 0x00, 0xab, 0xae,
		0x9f, 0x08, 0xf0, 0x99, 0xff, 0x9f, 0xdc, 0x2d,
		0x68, 0xd6, 0xe7, 0xe8, 0x44, 0x2a, 0x5b, 0x0e,
		0x48, 0x67, 0xe2, 0x41, 0x4a, 0xd9, 0x0c, 0x2a,
		0x2b, 0x4e, 0x66, 0x09, 0x87, 0xa0, 0x6b, 0x3b,
		0xd1, 0xd9, 0xa3, 0xe3, 0xa5, 0x69, 0xed, 0xc1,
		0x42, 0x03, 0x93, 0x0d, 0xbc, 0x7e, 0xe9, 0x08,
	}

	signature_1m_correct := [96]byte{
		0xf8, 0xb1, 0x20, 0xf2, 0x1e, 0x5c, 0xbf, 0x5f,
		0xea, 0x07, 0xcb, 0xb5, 0x77, 0xb8, 0x03, 0xbc,
		0xcb, 0x6d, 0xf1, 0xc1, 0xa5, 0x03, 0x05, 0x7b,
		0x01, 0x63, 0x9b, 0xf9, 0xed, 0x3e, 0x57, 0x47,
		0xd2, 0x5b, 0xf4, 0x7e, 0x7c, 0x45, 0xce, 0xfc,
		0x06, 0xb3, 0xf4, 0x05, 0x81, 0x9f, 0x53, 0xb0,
		0x18, 0xe3, 0xfa, 0xcb, 0xb2, 0x52, 0x3e, 0x57,
		0xcb, 0x34, 0xcc, 0x81, 0x60, 0xb9, 0x0b, 0x04,
		0x07, 0x79, 0xc0, 0x53, 0xad, 0xc4, 0x4b, 0xd0,
		0xb5, 0x7d, 0x95, 0x4e, 0xbe, 0xa5, 0x75, 0x0c,
		0xd4, 0xbf, 0xa7, 0xc0, 0xcf, 0xba, 0xe7, 0x7c,
		0xe2, 0x90, 0xef, 0x61, 0xa9, 0x29, 0x66, 0x0d,
	}

	signature_10m_correct := [96]byte{
		0xf5, 0xa4, 0xbc, 0xec, 0xc3, 0x3d, 0xd0, 0x43,
		0xd2, 0x81, 0x27, 0x9e, 0xf0, 0x4c, 0xbe, 0xf3,
		0x77, 0x01, 0x56, 0x41, 0x0e, 0xff, 0x0c, 0xb9,
		0x66, 0xec, 0x4d, 0xe0, 0xb7, 0x25, 0x63, 0x6b,
		0x5c, 0x08, 0x39, 0x80, 0x4e, 0x37, 0x1b, 0x2c,
		0x46, 0x6f, 0x86, 0x99, 0x1c, 0x4e, 0x31, 0x60,
		0xdb, 0x4c, 0xfe, 0xc5, 0xa2, 0x4d, 0x71, 0x2b,
		0xd6, 0xd0, 0xc3, 0x98, 0x88, 0xdb, 0x0e, 0x0c,
		0x68, 0x4a, 0xd3, 0xc7, 0x56, 0xac, 0x8d, 0x95,
		0x7b, 0xbd, 0x99, 0x50, 0xe8, 0xd3, 0xea, 0xf3,
		0x7b, 0x26, 0xf2, 0xa2, 0x2b, 0x02, 0x58, 0xca,
		0xbd, 0x2c, 0x2b, 0xf7, 0x77, 0x58, 0xfe, 0x09,
	}

	const msgLen = 200
	var privateKey [32]byte
	msg := make([]byte, msgLen)
	var random [64]byte
	signature := bytes.Repeat([]byte{3}, 96)

	t.Log("Pseudorandom VXEdDSA...\n")

	// up to 10000000 iterations (super slow ... )
	const iterations = 100000
	for count := 1; count <= iterations; count++ {
		b := sha512.Sum512(signature[:96])
		copy(privateKey[:], b[:32])
		b = sha512.Sum512(privateKey[:32])
		copy(random[:], b[:64])

		edwards25519.ScClamp(&privateKey)
		sk, err := GenerateKey(bytes.NewReader(privateKey[:]))
		if err != nil {
			t.Fatalf("Couldn't generate key in %d\n", count)
		}
		pk := sk.Public()
		signature = sk.signInternal(msg[:], bytes.NewReader(random[:]))
		if ok, _ := pk.verifyInteral(msg[:], signature); !ok {
			t.Fatalf("VXEdDSA verify failure #1 %d\n", count)
		}
		if (b[63] & 1) == 1 {
			signature[count%96] ^= 1
		} else {
			msg[count%msgLen] ^= 1
		}

		if count == 10000 {
			if !bytes.Equal(signature, signature_10k_correct[:]) {
				t.Errorf("VXEDDSA 10K doesn't match %d\n", count)
			}
		}
		if count == 100000 {
			if !bytes.Equal(signature, signature_100k_correct[:]) {
				t.Errorf("VXEDDSA 100K doesn't match %d\n", count)
			}
		}
		if count == 1000000 {
			if !bytes.Equal(signature, signature_1m_correct[:]) {
				t.Errorf("VXEDDSA 1m doesn't match %d\n", count)
			}
		}
		if count == 10000000 {
			if !bytes.Equal(signature, signature_10m_correct[:]) {
				t.Errorf("VXEDDSA 10m doesn't match %d\n", count)
			}
		}
	}
}

func BenchmarkCompute(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sk.Compute(alice)
	}
}

func BenchmarkProve(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sk.Prove(alice)
	}
}

func BenchmarkVerify(b *testing.B) {
	sk, err := GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	alice := []byte("alice")
	aliceVRF := sk.Compute(alice)
	_, aliceProof := sk.Prove(alice)
	pk := sk.Public()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pk.Verify(alice, aliceVRF, aliceProof)
	}
}
