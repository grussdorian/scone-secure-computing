package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	key1 := flag.String("key1", "", "32 byte AES key 1")
	key2 := flag.String("key2", "", "32 byte AES key 2")
	key3 := flag.String("key3", "", "32 byte AES key 3")
	flag.Parse()

	// Combine keys
	combinedKey := xorKeys(*key1, *key2, *key3)
	fmt.Printf("Combined key: %x\n", combinedKey)
	// Read file
	filename := "/volumes/v1/file.txt"
	plaintext, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nContents of the file before encryption:\n\n%s\n\n", string(plaintext))

	ciphertext, err := encrypt(plaintext, combinedKey)
	if err != nil {
		log.Fatal(err)
	}

	encryptedFilename := "../logs/file.aes"
	err = ioutil.WriteFile(encryptedFilename, ciphertext, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nEncrypted file written to %s\n\n", encryptedFilename)

	encryptedFile, err := os.Open(encryptedFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer encryptedFile.Close()

	decryptedFile, err := decrypt(encryptedFile, combinedKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Contents of the file after decryption:\n\n\n%s\n", string(decryptedFile))
}

func xorKeys(key1 string, key2 string, key3 string) []byte {

	key1Bytes, _ := hex.DecodeString(key1)
	key2Bytes, _ := hex.DecodeString(key2)
	key3Bytes, _ := hex.DecodeString(key3)
	// XOR keys together
	combinedKey := make([]byte, 32)
	for i := 0; i < 32; i++ {
		combinedKey[i] = key1Bytes[i] ^ key2Bytes[i] ^ key3Bytes[i]
	}
	return combinedKey
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Generate random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext[:aes.BlockSize], iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func decrypt(ciphertext io.Reader, key []byte) ([]byte, error) {
	// Read IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(ciphertext, iv); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, 0)
	stream := cipher.NewCFBDecrypter(block, iv)
	buf := make([]byte, 1024)
	for {
		n, err := ciphertext.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		stream.XORKeyStream(buf[:n], buf[:n])
		plaintext = append(plaintext, buf[:n]...)
	}

	return plaintext, nil
}
