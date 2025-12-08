package crypt

import (
	"testing"
)

func TestAESCrypt_Encrypt(t *testing.T) {
	type fields struct {
		secretKey []byte
	}
	type args struct {
		msg []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "#1",
			fields:  fields{secretKey: []byte("Y0UX8nx5PfZ04tpiY0UX8nx5PfZ04tpi")},
			args:    args{msg: []byte("hello there")},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AESCrypt{
				secretKey: tt.fields.secretKey,
			}
			_, err := a.Encrypt(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCrypt.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAESCrypt_Decrypt(t *testing.T) {
	type fields struct {
		secretKey []byte
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "#1",
			fields:  fields{secretKey: []byte("Y0UX8nx5PfZ04tpiY0UX8nx5PfZ04tpi")},
			args:    args{msg: "KlHxATnEG787v4r7nkKG+bXKkvZiZJQm7BhQ"},
			want:    "hello there",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AESCrypt{
				secretKey: tt.fields.secretKey,
			}
			got, err := a.Decrypt(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCrypt.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AESCrypt.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
