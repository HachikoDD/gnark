/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mimc

import (
	"testing"

	"github.com/consensys/gnark/backend/groth16"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gurvy"

	mimcbls377 "github.com/consensys/gnark/crypto/hash/mimc/bls377"
	mimcbls381 "github.com/consensys/gnark/crypto/hash/mimc/bls381"
	mimcbn256 "github.com/consensys/gnark/crypto/hash/mimc/bn256"

	fr_bls377 "github.com/consensys/gurvy/bls377/fr"
	fr_bls381 "github.com/consensys/gurvy/bls381/fr"
	fr_bn256 "github.com/consensys/gurvy/bn256/fr"
)

type mimcCircuit struct {
	ExpectedResult frontend.Variable `gnark:"ExpectedHash,public"`
	Data           frontend.Variable
}

func (circuit *mimcCircuit) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	result := mimc.Hash(cs, circuit.Data)
	cs.AssertIsEqual(result, circuit.ExpectedResult)
	return nil
}

func TestMimcBN256(t *testing.T) {
	assert := groth16.NewAssert(t)

	// input
	var data fr_bn256.Element
	data.SetString("7808462342289447506325013279997289618334122576263655295146895675168642919487")

	// minimal cs res = hash(data)
	var circuit, witness mimcCircuit
	r1cs, err := frontend.Compile(gurvy.BN256, &circuit)
	if err != nil {
		t.Fatal(err)

	}
	// running MiMC (Go)
	b := mimcbn256.Sum("seed", data.Bytes())
	var tmp fr_bn256.Element
	tmp.SetBytes(b)
	witness.Data.Assign(data)
	witness.ExpectedResult.Assign(tmp)

	// creates r1cs
	assert.SolvingSucceeded(r1cs, &witness)
}

func TestMimcBLS381(t *testing.T) {

	assert := groth16.NewAssert(t)

	// input
	var data fr_bls381.Element
	data.SetString("7808462342289447506325013279997289618334122576263655295146895675168642919487")

	// minimal cs res = hash(data)
	var circuit, witness mimcCircuit
	r1cs, err := frontend.Compile(gurvy.BLS381, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// running MiMC (Go)
	b := mimcbls381.Sum("seed", data.Bytes())
	var tmp fr_bls381.Element
	tmp.SetBytes(b)
	witness.Data.Assign(data)
	witness.ExpectedResult.Assign(tmp)

	assert.SolvingSucceeded(r1cs, &witness)

}

func TestMimcBLS377(t *testing.T) {

	assert := groth16.NewAssert(t)

	// input
	var data fr_bls377.Element
	data.SetString("7808462342289447506325013279997289618334122576263655295146895675168642919487")

	// minimal cs res = hash(data)
	var circuit, witness mimcCircuit
	r1cs, err := frontend.Compile(gurvy.BLS377, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// running MiMC (Go)
	b := mimcbls377.Sum("seed", data.Bytes())
	var tmp fr_bls377.Element
	tmp.SetBytes(b)
	witness.Data.Assign(data)
	witness.ExpectedResult.Assign(tmp)

	assert.SolvingSucceeded(r1cs, &witness)

}

//------------------------------------------------------------------
// benches

var nbHashes = [...]int{
	1 << 11,
	1 << 12,
	1 << 13,
	1 << 14,
	1 << 15,
	1 << 16,
	1 << 17,
	1 << 18,
}

// nb mimcs = 2**19
type batchMimc19 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc19) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[0]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc19(b *testing.B) {

	nbMimcs := nbHashes[0]

	var batchedMimc, witness batchMimc19
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**20
type batchMimc20 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc20) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[1]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc20(b *testing.B) {

	nbMimcs := nbHashes[1]

	var batchedMimc, witness batchMimc20
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**21
type batchMimc21 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc21) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[2]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc21(b *testing.B) {

	nbMimcs := nbHashes[2]

	var batchedMimc, witness batchMimc21
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**22
type batchMimc22 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc22) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[3]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc22(b *testing.B) {

	nbMimcs := nbHashes[3]

	var batchedMimc, witness batchMimc22
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**23
type batchMimc23 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc23) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[4]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc23(b *testing.B) {

	nbMimcs := nbHashes[4]

	var batchedMimc, witness batchMimc23
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**24
type batchMimc24 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc24) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[5]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc24(b *testing.B) {

	nbMimcs := nbHashes[5]

	var batchedMimc, witness batchMimc24
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}

// nb mimcs = 2**25
type batchMimc25 struct {
	ExpectedResult []frontend.Variable `gnark:"ExpectedHash,public"`
	Data           []frontend.Variable
}

func (batch *batchMimc25) Define(curveID gurvy.ID, cs *frontend.ConstraintSystem) error {
	mimc, err := NewMiMC("seed", curveID)
	if err != nil {
		return err
	}
	for i := 0; i < nbHashes[6]; i++ {
		result := mimc.Hash(cs, batch.Data[i])
		cs.AssertIsEqual(result, batch.ExpectedResult[i])
	}
	return nil
}

func BenchmarkBatchMimc25(b *testing.B) {

	nbMimcs := nbHashes[6]

	var batchedMimc, witness batchMimc25
	batchedMimc.ExpectedResult = make([]frontend.Variable, nbMimcs)
	batchedMimc.Data = make([]frontend.Variable, nbMimcs)
	witness.ExpectedResult = make([]frontend.Variable, nbMimcs)
	witness.Data = make([]frontend.Variable, nbMimcs)

	var sample fr_bn256.Element

	for j := 0; j < nbMimcs; j++ {
		sample.SetRandom() // so the multi exp is not trivial
		witness.ExpectedResult[j].Assign(sample)
		sample.SetRandom()
		witness.Data[j].Assign(sample)
	}

	r1cs, _ := frontend.Compile(gurvy.BN256, &batchedMimc)

	pk := groth16.DummySetup(r1cs)

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		groth16.ProveUnsafe(r1cs, pk, &witness) // <- the constraint system is not satisfied, but the full proving algo is performed
	}
}
