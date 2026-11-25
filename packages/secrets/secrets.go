package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
)

const (
	formatVersion = byte(1)
	keySize       = 32
)

var (
	ErrKeyring = errors.New("invalid encryption keyring")
	ErrSeal    = errors.New("unable to seal secret")
	ErrOpen    = errors.New("unable to open secret")
)

type Key struct {
	ID     string
	Secret [keySize]byte
}

type Keyring struct {
	active string
	keys   map[string][keySize]byte
}

func NewKeyring(keys []Key) (*Keyring, error) {
	if len(keys) == 0 {
		return nil, ErrKeyring
	}
	ring := &Keyring{active: keys[0].ID, keys: make(map[string][keySize]byte, len(keys))}
	for _, key := range keys {
		if key.ID == "" {
			return nil, ErrKeyring
		}
		if _, exists := ring.keys[key.ID]; exists {
			return nil, ErrKeyring
		}
		ring.keys[key.ID] = key.Secret
	}
	return ring, nil
}

func (k *Keyring) Seal(plaintext, aad []byte) ([]byte, string, error) {
	dek := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, "", ErrSeal
	}
	activeKey := k.keys[k.active]
	wrapAEAD, err := newAEAD(activeKey[:])
	if err != nil {
		return nil, "", ErrSeal
	}
	dataAEAD, err := newAEAD(dek)
	if err != nil {
		return nil, "", ErrSeal
	}
	wrapNonce := make([]byte, wrapAEAD.NonceSize())
	dataNonce := make([]byte, dataAEAD.NonceSize())
	if _, err := io.ReadFull(rand.Reader, wrapNonce); err != nil {
		return nil, "", ErrSeal
	}
	if _, err := io.ReadFull(rand.Reader, dataNonce); err != nil {
		return nil, "", ErrSeal
	}
	wrappedDEK := wrapAEAD.Seal(nil, wrapNonce, dek, []byte(k.active))
	sealedData := dataAEAD.Seal(nil, dataNonce, plaintext, aad)
	ciphertext := make([]byte, 1, 1+len(wrapNonce)+len(wrappedDEK)+len(dataNonce)+len(sealedData))
	ciphertext[0] = formatVersion
	ciphertext = append(ciphertext, wrapNonce...)
	ciphertext = append(ciphertext, wrappedDEK...)
	ciphertext = append(ciphertext, dataNonce...)
	ciphertext = append(ciphertext, sealedData...)
	return ciphertext, k.active, nil
}

func (k *Keyring) Open(ciphertext []byte, keyID string, aad []byte) ([]byte, error) {
	const minimumSize = 1 + 12 + keySize + 16 + 12 + 16
	if len(ciphertext) < minimumSize || ciphertext[0] != formatVersion {
		return nil, ErrOpen
	}
	kek, ok := k.keys[keyID]
	if !ok {
		return nil, ErrOpen
	}
	wrapAEAD, err := newAEAD(kek[:])
	if err != nil {
		return nil, ErrOpen
	}
	wrapNonce := ciphertext[1:13]
	wrappedDEK := ciphertext[13:61]
	dataNonce := ciphertext[61:73]
	wrappedData := ciphertext[73:]
	dek, err := wrapAEAD.Open(nil, wrapNonce, wrappedDEK, []byte(keyID))
	if err != nil {
		return nil, ErrOpen
	}
	dataAEAD, err := newAEAD(dek)
	if err != nil {
		return nil, ErrOpen
	}
	plaintext, err := dataAEAD.Open(nil, dataNonce, wrappedData, aad)
	if err != nil {
		return nil, ErrOpen
	}
	return plaintext, nil
}

func RecordAAD(table, column, organizationID, recordID string) []byte {
	fields := [...]string{table, column, organizationID, recordID}
	var size int
	for _, field := range fields {
		size += 8 + len(field)
	}
	out := make([]byte, size)
	offset := 0
	for _, field := range fields {
		binary.BigEndian.PutUint64(out[offset:offset+8], uint64(len(field)))
		offset += 8
		copy(out[offset:], field)
		offset += len(field)
	}
	return out
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
