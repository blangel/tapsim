package script

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/halseth/tapsim/output"
	"github.com/pkg/term"
)

const scriptFlags = txscript.StandardVerifyFlags

// Execute builds a tap leaf using the passed pkScript and executes it step by
// step with the provided witness.
//
// privKeyBytes should map names of private keys given in the input witness to
// key bytes. An empty key will generate a random one.
//
// If [input/output]KeyBytes is empty, a random key will be generated.
func Execute(privKeyBytes map[string][]byte, inputKeyBytes, outputKeyBytes []byte, pkScript []byte,
	witnessGen []WitnessGen, interactive bool, tags map[string]string) error {

	// Parse the input private keys.
	privKeys := make(map[string]*btcec.PrivateKey)
	for k, v := range privKeyBytes {
		var (
			key *btcec.PrivateKey
			err error
		)
		// If the key is empty, generate a random one.
		if len(v) == 0 {
			key, err = btcec.NewPrivateKey()
			if err != nil {
				return err
			}
		} else {
			key, _ = btcec.PrivKeyFromBytes(v)
		}
		privKeys[k] = key
	}

	var inputKey *btcec.PublicKey
	if len(inputKeyBytes) == 0 {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}

		inputKey = privKey.PubKey()
	} else {
		var err error
		inputKey, err = schnorr.ParsePubKey(inputKeyBytes)
		if err != nil {
			return err
		}
	}

	var outputKey *btcec.PublicKey
	if len(outputKeyBytes) == 0 {
		privKey, err := btcec.NewPrivateKey()
		if err != nil {
			return err
		}

		outputKey = privKey.PubKey()
	} else {
		var err error
		outputKey, err = schnorr.ParsePubKey(outputKeyBytes)
		if err != nil {
			return err
		}
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
		inputKey,
	)

	tapScriptRootHash := tapScriptTree.RootNode.TapHash()

	inputTapKey := txscript.ComputeTaprootOutputKey(
		inputKey, tapScriptRootHash[:],
	)

	inputScript, err := txscript.PayToTaprootScript(inputTapKey)
	if err != nil {
		return err
	}

	outputTapKey := txscript.ComputeTaprootOutputKey(
		outputKey, tapScriptRootHash[:],
	)

	fmt.Printf("taptree: %x\n", tapScriptRootHash[:])
	fmt.Printf("input internal key: %x\n", schnorr.SerializePubKey(inputKey))
	fmt.Printf("input taproot key: %x\n", schnorr.SerializePubKey(inputTapKey))
	fmt.Printf("output internal key: %x\n", schnorr.SerializePubKey(outputKey))
	fmt.Printf("output taproot key: %x\n", schnorr.SerializePubKey(outputTapKey))
	outputScript, err := txscript.PayToTaprootScript(outputTapKey)
	if err != nil {
		return err
	}

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Index: 0,
		},
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    1e8,
		PkScript: outputScript,
	})

	prevOut := &wire.TxOut{
		Value:    1e8,
		PkScript: inputScript,
	}
	prevOutFetcher := txscript.NewCannedPrevOutputFetcher(
		prevOut.PkScript, prevOut.Value,
	)

	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	signFunc := func(keyID string) ([]byte, error) {
		privKey, ok := privKeys[keyID]
		if !ok {
			return nil, fmt.Errorf("private key %s not known", keyID)
		}
		return txscript.RawTxInTapscriptSignature(
			tx, sigHashes, 0, prevOut.Value, prevOut.PkScript, tapLeaf,
			txscript.SigHashDefault, privKey,
		)
	}

	var combinedWitness wire.TxWitness
	for _, gen := range witnessGen {
		w, err := gen(signFunc)
		if err != nil {
			return err
		}

		combinedWitness = append(combinedWitness, w)
	}

	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return err
	}

	combinedWitness = append(combinedWitness, pkScript, ctrlBlockBytes)

	txCopy := tx.Copy()
	txCopy.TxIn[0].Witness = wire.TxWitness{}
	txCopy.TxIn[0].Witness = combinedWitness

	setupFunc := func() (*txscript.Engine, error) {
		sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
		return txscript.NewEngine(
			prevOut.PkScript, txCopy, 0, scriptFlags,
			nil, sigHashes, prevOut.Value, prevOutFetcher,
		)
	}

	var t *term.Term
	if interactive {
		// Set the terminal in raw mode, such that we can capture arrow
		// presses.
		t, err = term.Open("/dev/tty")
		if err != nil {
			return err
		}
		defer t.Close()

		term.RawMode(t)
		defer t.Restore()
	}

	currentStep := 1
	prevLines := 0
	bytes := make([]byte, 3)
	for {
		// Based on the current step counter, we execute up until that
		// step, then print the state table.
		table, vmErr := StepScript(
			setupFunc, txCopy.TxIn[0].Witness, tags, currentStep,
		)

		// Before handling any error, we draw the state table for the
		// step.
		clearLines := 0
		if interactive {
			clearLines = prevLines
		}

		output.DrawTable(table, clearLines)
		if interactive {
			if currentStep > 1 {
				fmt.Printf("Script execution: \u2190 back | next \u2192 ")
			} else {
				fmt.Printf("Script execution: next \u2192 ")
			}
		}

		// Take note of the number of lines just printed, such that we
		// can clear them on next iteration in case we are using
		// interactive mode.
		prevLines = strings.Count(table, "\n") + 1

		// If the VM encountered no error, it means the script
		// successfully executed to completion.
		if vmErr == nil {
			output.ClearLines(1)
			return nil
		}

		// If we encountered an error other than errAbortVM,
		// the script actually failed.
		if vmErr != errAbortVM {
			output.ClearLines(1)
			return vmErr
		}

		// Otherwise script execution was aborted before it completed,
		// so we continue with the next step of the execution.

		if interactive {
			numRead, err := t.Read(bytes)
			if err != nil {
				return err
			}

			if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
				switch bytes[2] {
				case 65:
					//fmt.Print("Up arrow key pressed\r\n")
				case 66:
					//fmt.Print("Down arrow key pressed\r\n")
				case 67:
					//fmt.Print("Right arrow key pressed\r\n")
					currentStep++
				case 68:
					//fmt.Print("Left arrow key pressed\r\n")
					currentStep--
					if currentStep < 1 {
						currentStep = 1
					}
				}

			} else if numRead == 1 && bytes[0] == 3 {
				// Ctrl+C pressed, quit the program
				output.ClearLines(1)
				return fmt.Errorf("execution aborted")
			}
		} else {
			currentStep++
		}
	}
}

var errAbortVM = fmt.Errorf("aborting vm execution")

func StepScript(setupFunc func() (*txscript.Engine, error), witness [][]byte,
	tags map[string]string, numSteps int) (string, error) {

	vm, err := setupFunc()
	if err != nil {
		return "", err
	}

	const (
		SCRIPT_SCRIPTSIG      = 0
		SCRIPT_SCRIPTPUBKEY   = 1
		SCRIPT_WITNESS_SCRIPT = 2
	)

	// Set up a callback that we will use to inspect the engine state at
	// every execution step.
	var (
		currentScript = -1
		stepCounter   = 0
		finalState    string
	)
	vm.StepCallback = func(step *txscript.StepInfo) error {
		finalState = ""
		var showWitness [][]byte

		switch step.ScriptIndex {
		// Script sig is empty and uninteresting under segwit, so we
		// just ignore it.
		case SCRIPT_SCRIPTSIG:
			currentScript = step.ScriptIndex
			return nil

		// The scriptpubkey contains the witness program and is used to
		// verify the script in the provided witness.
		case SCRIPT_SCRIPTPUBKEY:
			// Since to real script execution is done during the
			// script pubkey (only checking the witness program),
			// we will only output the step the first time we
			// encounter this script index.
			if currentScript == step.ScriptIndex {
				return nil
			}

			showWitness = witness

		// Execution of the witness script is the interesting part.
		case SCRIPT_WITNESS_SCRIPT:
			if currentScript != step.ScriptIndex {
				finalState += "witness program verified OK\n"
			}
		}

		stepCounter++
		currentScript = step.ScriptIndex

		// Parse the current script for output.
		scriptStr := output.VmScriptToString(vm, step.ScriptIndex)
		table := output.ExecutionTable(
			step.OpcodeIndex,
			scriptStr,
			output.StackToString(step.Stack),
			output.StackToString(step.AltStack),
			output.StackToString(showWitness),
			tags,
		)

		finalState += table
		finalState += "\n"

		// If we have executed enough steps, signal the VM to abort
		// using our custom error.
		if stepCounter >= numSteps {
			return errAbortVM
		}

		return nil
	}

	vmErr := vm.Execute()
	return finalState, vmErr
}
