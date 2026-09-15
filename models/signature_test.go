package models

import (
	"encoding/base64"
	"testing"
)

func TestFromSerializedSignatureMultiSig(t *testing.T) {
	serialized := base64.StdEncoding.EncodeToString([]byte{3, 1, 2, 3})
	parsed, err := FromSerializedSignature(serialized)
	if err != nil {
		t.Fatalf("FromSerializedSignature() error = %v", err)
	}
	if parsed.SignatureScheme != "MultiSig" {
		t.Fatalf("signature scheme = %s, want MultiSig", parsed.SignatureScheme)
	}
	if got, want := len(parsed.Signature), 3; got != want {
		t.Fatalf("multisig payload length = %d, want %d", got, want)
	}
}

func TestVerifyPersonalMessage(t *testing.T) {
	type args struct {
		message   string
		signature string
	}
	tests := []struct {
		name       string
		args       args
		wantSigner string
		wantPass   bool
		wantErr    bool
	}{
		{
			name: "Test 1",
			args: args{
				message:   "123456 is the thing that you need to sign",
				signature: "AIjj13rXd9GFZRNPd4XNUvthHMHg5bovf8/mW4a7EYAWC6mQtAAaa0tSPhk6YpNED34/qeaCYwnN1QAsKm253gfQ6i6fULpM+uscFuJIXoTT/JQvMo3CUlLODcGxPkUbHg==",
			},
			wantSigner: "0x00dccd645260cfe9145bdabb7b45b42e188af8661086aa7bb2e7f3adc1cd2785",
			wantPass:   true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSigner, gotPass, err := VerifyPersonalMessage(tt.args.message, tt.args.signature)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPersonalMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSigner != tt.wantSigner {
				t.Errorf("VerifyPersonalMessage() gotSigner = %v, want %v", gotSigner, tt.wantSigner)
			}
			if gotPass != tt.wantPass {
				t.Errorf("VerifyPersonalMessage() gotPass = %v, want %v", gotPass, tt.wantPass)
			}
		})
	}
}
