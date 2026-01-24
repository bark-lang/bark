package modules

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"gitlab.com/bark-lang/bark/evaluator/builtins/helpers"
	"gitlab.com/bark-lang/bark/object"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// InitCrypto initializes cryptography operations
func InitCrypto() map[string]*object.Builtin {
	return map[string]*object.Builtin{
		// Password Hashing - bcrypt
		"crypto.bcrypt_hash": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.bcrypt_hash requires 1 argument (password), got=%d", len(args))
				}

				password, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.bcrypt_hash requires string argument, got=%s", args[0].Type())
				}

				// Use cost 12 (recommended default, ~250ms on modern hardware)
				hash, err := bcrypt.GenerateFromPassword([]byte(password.Value), 12)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("bcrypt hashing failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: string(hash)},
					},
				}
			},
		},

		"crypto.bcrypt_verify": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.bcrypt_verify requires 2 arguments (password, hash), got=%d", len(args))
				}

				password, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.bcrypt_verify requires string password, got=%s", args[0].Type())
				}

				hash, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.bcrypt_verify requires string hash, got=%s", args[1].Type())
				}

				err := bcrypt.CompareHashAndPassword([]byte(hash.Value), []byte(password.Value))
				if err != nil {
					// Wrong password or invalid hash
					return helpers.FALSE
				}

				return helpers.TRUE
			},
		},

		// Password Hashing - argon2id (more secure, recommended for new systems)
		"crypto.argon2_hash": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.argon2_hash requires 1 argument (password), got=%d", len(args))
				}

				password, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.argon2_hash requires string argument, got=%s", args[0].Type())
				}

				// Generate 16-byte salt
				salt := make([]byte, 16)
				if _, err := io.ReadFull(rand.Reader, salt); err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("salt generation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// OWASP recommendations: time=2, memory=19456 (19MB), threads=1
				hash := argon2.IDKey([]byte(password.Value), salt, 2, 19*1024, 1, 32)

				// Format: $argon2id$salt$hash (base64 encoded)
				encoded := fmt.Sprintf("$argon2id$%s$%s",
					base64.RawStdEncoding.EncodeToString(salt),
					base64.RawStdEncoding.EncodeToString(hash))

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: encoded},
					},
				}
			},
		},

		"crypto.argon2_verify": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.argon2_verify requires 2 arguments (password, hash), got=%d", len(args))
				}

				password, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.argon2_verify requires string password, got=%s", args[0].Type())
				}

				encoded, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.argon2_verify requires string hash, got=%s", args[1].Type())
				}

				// Parse format: $argon2id$salt$hash
				// Use strings.Split for more reliable parsing
				parts := strings.Split(encoded.Value, "$")
				// Expected: ["", "argon2id", "salt", "hash"]
				if len(parts) != 4 || parts[0] != "" || parts[1] != "argon2id" {
					return helpers.FALSE // Invalid format
				}

				salt, err := base64.RawStdEncoding.DecodeString(parts[2])
				if err != nil {
					return helpers.FALSE
				}

				expectedHash, err := base64.RawStdEncoding.DecodeString(parts[3])
				if err != nil {
					return helpers.FALSE
				}

				// Re-hash with same parameters
				hash := argon2.IDKey([]byte(password.Value), salt, 2, 19*1024, 1, 32)

				// Constant-time comparison
				if hmac.Equal(hash, expectedHash) {
					return helpers.TRUE
				}

				return helpers.FALSE
			},
		},

		// AES Encryption/Decryption (AES-256-GCM - authenticated encryption)
		"crypto.aes_encrypt": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.aes_encrypt requires 2 arguments (plaintext, key), got=%d", len(args))
				}

				plaintext, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.aes_encrypt requires string plaintext, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.aes_encrypt requires string key, got=%s", args[1].Type())
				}

				// Key must be 32 bytes for AES-256
				if len(key.Value) != 32 {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "key must be exactly 32 bytes for AES-256",
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				block, err := aes.NewCipher([]byte(key.Value))
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("cipher creation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				gcm, err := cipher.NewGCM(block)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("GCM mode failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// Generate random nonce
				nonce := make([]byte, gcm.NonceSize())
				if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("nonce generation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// Encrypt and authenticate (nonce is prepended to ciphertext)
				ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext.Value), nil)

				// Return hex-encoded ciphertext
				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: hex.EncodeToString(ciphertext)},
					},
				}
			},
		},

		"crypto.aes_decrypt": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.aes_decrypt requires 2 arguments (ciphertext, key), got=%d", len(args))
				}

				ciphertextHex, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.aes_decrypt requires string ciphertext, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.aes_decrypt requires string key, got=%s", args[1].Type())
				}

				// Key must be 32 bytes for AES-256
				if len(key.Value) != 32 {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "key must be exactly 32 bytes for AES-256",
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// Decode hex ciphertext
				ciphertext, err := hex.DecodeString(ciphertextHex.Value)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("invalid ciphertext encoding: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				block, err := aes.NewCipher([]byte(key.Value))
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("cipher creation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				gcm, err := cipher.NewGCM(block)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("GCM mode failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				nonceSize := gcm.NonceSize()
				if len(ciphertext) < nonceSize {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "ciphertext too short",
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
				plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
				if err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("decryption failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: string(plaintext)},
					},
				}
			},
		},

		// HMAC Signatures
		"crypto.hmac_sha256": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.hmac_sha256 requires 2 arguments (message, key), got=%d", len(args))
				}

				message, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_sha256 requires string message, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_sha256 requires string key, got=%s", args[1].Type())
				}

				h := hmac.New(sha256.New, []byte(key.Value))
				h.Write([]byte(message.Value))
				signature := h.Sum(nil)

				return &object.String{Value: hex.EncodeToString(signature)}
			},
		},

		"crypto.hmac_sha512": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return helpers.NewError("crypto.hmac_sha512 requires 2 arguments (message, key), got=%d", len(args))
				}

				message, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_sha512 requires string message, got=%s", args[0].Type())
				}

				key, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_sha512 requires string key, got=%s", args[1].Type())
				}

				h := hmac.New(sha512.New, []byte(key.Value))
				h.Write([]byte(message.Value))
				signature := h.Sum(nil)

				return &object.String{Value: hex.EncodeToString(signature)}
			},
		},

		"crypto.hmac_verify": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return helpers.NewError("crypto.hmac_verify requires 3 arguments (message, signature, key), got=%d", len(args))
				}

				message, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_verify requires string message, got=%s", args[0].Type())
				}

				signatureHex, ok := args[1].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_verify requires string signature, got=%s", args[1].Type())
				}

				key, ok := args[2].(*object.String)
				if !ok {
					return helpers.NewError("crypto.hmac_verify requires string key, got=%s", args[2].Type())
				}

				// Decode expected signature
				expectedSig, err := hex.DecodeString(signatureHex.Value)
				if err != nil {
					return helpers.FALSE // Invalid signature format
				}

				// Try SHA-256 first (32 bytes)
				if len(expectedSig) == 32 {
					h := hmac.New(sha256.New, []byte(key.Value))
					h.Write([]byte(message.Value))
					computedSig := h.Sum(nil)

					if hmac.Equal(computedSig, expectedSig) {
						return helpers.TRUE
					}
					return helpers.FALSE
				}

				// Try SHA-512 (64 bytes)
				if len(expectedSig) == 64 {
					h := hmac.New(sha512.New, []byte(key.Value))
					h.Write([]byte(message.Value))
					computedSig := h.Sum(nil)

					if hmac.Equal(computedSig, expectedSig) {
						return helpers.TRUE
					}
					return helpers.FALSE
				}

				// Unknown signature length
				return helpers.FALSE
			},
		},

		// Secure Random Generation
		"crypto.random_bytes": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.random_bytes requires 1 argument (length), got=%d", len(args))
				}

				var length int
				switch v := args[0].(type) {
				case *object.Integer:
					length = int(v.Value)
				case *object.Float:
					length = int(v.Value)
				default:
					return helpers.NewError("crypto.random_bytes requires number argument, got=%s", args[0].Type())
				}

				if length < 0 || length > 1024 {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "length must be between 0 and 1024",
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				bytes := make([]byte, length)
				if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("random generation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: hex.EncodeToString(bytes)},
					},
				}
			},
		},

		"crypto.random_string": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.random_string requires 1 argument (length), got=%d", len(args))
				}

				var length int
				switch v := args[0].(type) {
				case *object.Integer:
					length = int(v.Value)
				case *object.Float:
					length = int(v.Value)
				default:
					return helpers.NewError("crypto.random_string requires number argument, got=%s", args[0].Type())
				}

				if length < 0 || length > 1024 {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     "length must be between 0 and 1024",
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// Use base64 URL-safe encoding for readable strings
				bytes := make([]byte, length)
				if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
					return &object.Tuple{
						Elements: []object.Object{
							&object.Error{
								Msg:     fmt.Sprintf("random generation failed: %s", err.Error()),
								Context: make(map[string]object.Object),
							},
							&object.String{Value: ""},
						},
					}
				}

				// Generate slightly more data to ensure we get desired length after encoding
				str := base64.URLEncoding.EncodeToString(bytes)[:length]

				return &object.Tuple{
					Elements: []object.Object{
						&object.Map{Pairs: make(map[string]object.Object), Keys: []string{}}, // null/empty error
						&object.String{Value: str},
					},
				}
			},
		},

		// Hash Functions (non-keyed, for checksums/fingerprints)
		"crypto.sha256": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.sha256 requires 1 argument (data), got=%d", len(args))
				}

				data, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.sha256 requires string argument, got=%s", args[0].Type())
				}

				hash := sha256.Sum256([]byte(data.Value))
				return &object.String{Value: hex.EncodeToString(hash[:])}
			},
		},

		"crypto.sha512": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return helpers.NewError("crypto.sha512 requires 1 argument (data), got=%d", len(args))
				}

				data, ok := args[0].(*object.String)
				if !ok {
					return helpers.NewError("crypto.sha512 requires string argument, got=%s", args[0].Type())
				}

				hash := sha512.Sum512([]byte(data.Value))
				return &object.String{Value: hex.EncodeToString(hash[:])}
			},
		},
	}
}
