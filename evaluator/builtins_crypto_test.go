package evaluator

import (
	"strings"
	"testing"

	"gitlab.com/bark-lang/barki/object"
)

func TestCryptoBcryptHash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.bcrypt_hash("password123")`, "array"},
		{`crypto.bcrypt_hash("test")`, "array"},
		{`crypto.bcrypt_hash("")`, "array"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected tuple, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements in arrayay, got=%d", len(arr.Elements))
			continue
		}

		// First element should be empty map (no error)
		if arr.Elements[0].Type() != object.MAP_OBJ {
			t.Errorf("expected map for no-error, got=%s", arr.Elements[0].Type())
		}

		// Second element should be non-empty hash string
		hash, ok := arr.Elements[1].(*object.String)
		if !ok {
			t.Errorf("expected string hash, got=%T", arr.Elements[1])
			continue
		}

		if len(hash.Value) == 0 {
			t.Errorf("expected non-empty hash")
		}

		// Bcrypt hashes start with $2a$ or $2b$
		if !strings.HasPrefix(hash.Value, "$2a$") && !strings.HasPrefix(hash.Value, "$2b$") {
			t.Errorf("expected bcrypt hash format, got=%s", hash.Value[:10])
		}
	}
}

func TestCryptoBcryptVerify(t *testing.T) {
	// First generate a hash
	hashResult := testEval(`crypto.bcrypt_hash("mypassword")`)
	arr := hashResult.(*object.Tuple)
	hash := arr.Elements[1].(*object.String)

	tests := []struct {
		input    string
		expected bool
	}{
		{`crypto.bcrypt_verify("mypassword", "` + hash.Value + `")`, true},
		{`crypto.bcrypt_verify("wrongpassword", "` + hash.Value + `")`, false},
		{`crypto.bcrypt_verify("", "` + hash.Value + `")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestCryptoBcryptErrors(t *testing.T) {
	tests := []string{
		`crypto.bcrypt_hash()`,           // wrong args
		`crypto.bcrypt_hash(123)`,        // wrong type
		`crypto.bcrypt_verify("pw")`,     // wrong args
		`crypto.bcrypt_verify(123, "h")`, // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Errorf("expected error for %s, got=%T", tt, evaluated)
		}
	}
}

func TestCryptoArgon2Hash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.argon2_hash("password123")`, "arr"},
		{`crypto.argon2_hash("test")`, "arr"},
		{`crypto.argon2_hash("")`, "arr"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected array, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements in array, got=%d", len(arr.Elements))
			continue
		}

		// First element should be null (no error)
		if arr.Elements[0].Type() != object.MAP_OBJ {
			t.Errorf("expected null for error, got=%s", arr.Elements[0].Type())
		}

		// Second element should be non-empty hash string
		hash, ok := arr.Elements[1].(*object.String)
		if !ok {
			t.Errorf("expected string hash, got=%T", arr.Elements[1])
			continue
		}

		if len(hash.Value) == 0 {
			t.Errorf("expected non-empty hash")
		}

		// Argon2 hashes start with $argon2id$
		if !strings.HasPrefix(hash.Value, "$argon2id$") {
			t.Errorf("expected argon2id hash format, got=%s", hash.Value[:20])
		}
	}
}

func TestCryptoArgon2Verify(t *testing.T) {
	// First generate a hash
	hashResult := testEval(`crypto.argon2_hash("mypassword")`)
	arr := hashResult.(*object.Tuple)
	hash := arr.Elements[1].(*object.String)

	tests := []struct {
		input    string
		expected bool
	}{
		{`crypto.argon2_verify("mypassword", "` + hash.Value + `")`, true},
		{`crypto.argon2_verify("wrongpassword", "` + hash.Value + `")`, false},
		{`crypto.argon2_verify("", "` + hash.Value + `")`, false},
		{`crypto.argon2_verify("test", "$invalid$format")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestCryptoArgon2Errors(t *testing.T) {
	tests := []string{
		`crypto.argon2_hash()`,           // wrong args
		`crypto.argon2_hash(123)`,        // wrong type
		`crypto.argon2_verify("pw")`,     // wrong args
		`crypto.argon2_verify(123, "h")`, // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Errorf("expected error for %s, got=%T", tt, evaluated)
		}
	}
}

func TestCryptoAESEncrypt(t *testing.T) {
	// Key must be exactly 32 bytes for AES-256
	key32 := "12345678901234567890123456789012"

	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.aes_encrypt("hello world", "` + key32 + `")`, "arr"},
		{`crypto.aes_encrypt("", "` + key32 + `")`, "arr"},
		{`crypto.aes_encrypt("sensitive data", "` + key32 + `")`, "arr"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected array, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements in array, got=%d", len(arr.Elements))
			continue
		}

		// First element should be null (no error)
		if arr.Elements[0].Type() != object.MAP_OBJ {
			errObj := arr.Elements[0].(*object.Error)
			t.Errorf("expected null for error, got error: %s", errObj.Msg)
			continue
		}

		// Second element should be non-empty hex string
		ciphertext, ok := arr.Elements[1].(*object.String)
		if !ok {
			t.Errorf("expected string ciphertext, got=%T", arr.Elements[1])
			continue
		}

		if len(ciphertext.Value) == 0 {
			t.Errorf("expected non-empty ciphertext")
		}
	}
}

func TestCryptoAESRoundTrip(t *testing.T) {
	key32 := "12345678901234567890123456789012"

	// Encrypt
	encryptResult := testEval(`crypto.aes_encrypt("secret message", "` + key32 + `")`)
	encryptTuple := encryptResult.(*object.Tuple)
	ciphertext := encryptTuple.Elements[1].(*object.String)

	// Decrypt
	decryptResult := testEval(`crypto.aes_decrypt("` + ciphertext.Value + `", "` + key32 + `")`)
	decryptTuple, ok := decryptResult.(*object.Tuple)
	if !ok {
		t.Fatalf("expected array from decrypt, got=%T", decryptResult)
	}

	// Check no error
	if decryptTuple.Elements[0].Type() != object.MAP_OBJ {
		errObj := decryptTuple.Elements[0].(*object.Error)
		t.Fatalf("decrypt failed: %s", errObj.Msg)
	}

	// Check plaintext matches
	plaintext := decryptTuple.Elements[1].(*object.String)
	if plaintext.Value != "secret message" {
		t.Errorf("expected 'secret message', got='%s'", plaintext.Value)
	}
}

func TestCryptoAESErrors(t *testing.T) {
	key32 := "12345678901234567890123456789012"
	keyShort := "short"

	tests := []struct {
		input       string
		expectError bool
	}{
		{`crypto.aes_encrypt()`, true},                             // wrong args
		{`crypto.aes_encrypt("data", "` + keyShort + `")`, true},   // key too short
		{`crypto.aes_decrypt("abc", "` + key32 + `")`, true},       // invalid hex
		{`crypto.aes_decrypt("deadbeef", "` + key32 + `")`, true},  // too short ciphertext
		{`crypto.aes_decrypt("aabbcc", "` + keyShort + `")`, true}, // key too short
		{`crypto.aes_encrypt(123, "` + key32 + `")`, true},         // wrong type
		{`crypto.aes_decrypt("data", 123)`, true},                  // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if tt.expectError {
			// Should be either Error or Tuple with Error as first element
			switch obj := evaluated.(type) {
			case *object.Error:
				// OK
			case *object.Tuple:
				if obj.Elements[0].Type() != object.ERROR_OBJ {
					t.Errorf("expected error in array for %s, got=%T", tt.input, obj.Elements[0])
				}
			default:
				t.Errorf("expected error for %s, got=%T", tt.input, evaluated)
			}
		}
	}
}

func TestCryptoHMACSHA256(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.hmac_sha256("message", "secret")`, "string"},
		{`crypto.hmac_sha256("", "key")`, "string"},
		{`crypto.hmac_sha256("data", "")`, "string"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		// SHA-256 produces 32 bytes = 64 hex characters
		if len(str.Value) != 64 {
			t.Errorf("expected 64 hex chars for SHA-256, got=%d", len(str.Value))
		}
	}
}

func TestCryptoHMACSHA512(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.hmac_sha512("message", "secret")`, "string"},
		{`crypto.hmac_sha512("", "key")`, "string"},
		{`crypto.hmac_sha512("data", "")`, "string"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		// SHA-512 produces 64 bytes = 128 hex characters
		if len(str.Value) != 128 {
			t.Errorf("expected 128 hex chars for SHA-512, got=%d", len(str.Value))
		}
	}
}

func TestCryptoHMACVerify(t *testing.T) {
	// Generate HMAC signatures
	sig256Result := testEval(`crypto.hmac_sha256("message", "secret")`)
	sig256 := sig256Result.(*object.String)

	sig512Result := testEval(`crypto.hmac_sha512("message", "secret")`)
	sig512 := sig512Result.(*object.String)

	tests := []struct {
		input    string
		expected bool
	}{
		{`crypto.hmac_verify("message", "` + sig256.Value + `", "secret")`, true},
		{`crypto.hmac_verify("message", "` + sig512.Value + `", "secret")`, true},
		{`crypto.hmac_verify("different", "` + sig256.Value + `", "secret")`, false},
		{`crypto.hmac_verify("message", "` + sig256.Value + `", "wrongkey")`, false},
		{`crypto.hmac_verify("message", "invalid", "secret")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestCryptoHMACErrors(t *testing.T) {
	tests := []string{
		`crypto.hmac_sha256()`,              // wrong args
		`crypto.hmac_sha256("msg")`,         // wrong args
		`crypto.hmac_sha256(123, "key")`,    // wrong type
		`crypto.hmac_sha512()`,              // wrong args
		`crypto.hmac_sha512("msg")`,         // wrong args
		`crypto.hmac_sha512(123, "key")`,    // wrong type
		`crypto.hmac_verify("msg", "sig")`,  // wrong args
		`crypto.hmac_verify(123, "s", "k")`, // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Errorf("expected error for %s, got=%T", tt, evaluated)
		}
	}
}

func TestCryptoRandomBytes(t *testing.T) {
	tests := []struct {
		input        string
		expectLength int
	}{
		{`crypto.random_bytes(16)`, 32}, // 16 bytes = 32 hex chars
		{`crypto.random_bytes(32)`, 64}, // 32 bytes = 64 hex chars
		{`crypto.random_bytes(0)`, 0},   // 0 bytes = 0 hex chars
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected array, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements in array, got=%d", len(arr.Elements))
			continue
		}

		// First element should be null (no error)
		if arr.Elements[0].Type() != object.MAP_OBJ {
			t.Errorf("expected null for error, got=%s", arr.Elements[0].Type())
		}

		// Second element should be hex string of correct length
		randomHex, ok := arr.Elements[1].(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T", arr.Elements[1])
			continue
		}

		if len(randomHex.Value) != tt.expectLength {
			t.Errorf("expected %d hex chars, got=%d", tt.expectLength, len(randomHex.Value))
		}
	}
}

func TestCryptoRandomString(t *testing.T) {
	tests := []struct {
		input        string
		expectLength int
	}{
		{`crypto.random_string(16)`, 16},
		{`crypto.random_string(32)`, 32},
		{`crypto.random_string(10)`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		arr, ok := evaluated.(*object.Tuple)
		if !ok {
			t.Errorf("expected array, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if len(arr.Elements) != 2 {
			t.Errorf("expected 2 elements in array, got=%d", len(arr.Elements))
			continue
		}

		// First element should be null (no error)
		if arr.Elements[0].Type() != object.MAP_OBJ {
			t.Errorf("expected null for error, got=%s", arr.Elements[0].Type())
		}

		// Second element should be string of correct length
		randomStr, ok := arr.Elements[1].(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T", arr.Elements[1])
			continue
		}

		if len(randomStr.Value) != tt.expectLength {
			t.Errorf("expected %d chars, got=%d", tt.expectLength, len(randomStr.Value))
		}
	}
}

func TestCryptoRandomErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{`crypto.random_bytes()`, true},      // wrong args
		{`crypto.random_bytes(-1)`, true},    // negative length
		{`crypto.random_bytes(2000)`, true},  // too large
		{`crypto.random_bytes("16")`, true},  // wrong type
		{`crypto.random_string()`, true},     // wrong args
		{`crypto.random_string(-1)`, true},   // negative length
		{`crypto.random_string(2000)`, true}, // too large
		{`crypto.random_string("16")`, true}, // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if tt.expectError {
			// Should be either Error or Tuple with Error as first element
			switch obj := evaluated.(type) {
			case *object.Error:
				// OK
			case *object.Tuple:
				if obj.Elements[0].Type() != object.ERROR_OBJ {
					t.Errorf("expected error in array for %s, got=%T", tt.input, obj.Elements[0])
				}
			default:
				t.Errorf("expected error for %s, got=%T", tt.input, evaluated)
			}
		}
	}
}

func TestCryptoSHA256(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.sha256("hello")`, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{`crypto.sha256("")`, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{`crypto.sha256("test")`, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if str.Value != tt.expected {
			t.Errorf("expected %s, got=%s", tt.expected, str.Value)
		}
	}
}

func TestCryptoSHA512(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`crypto.sha512("hello")`, "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
		{`crypto.sha512("")`, "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("expected string, got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if str.Value != tt.expected {
			t.Errorf("expected %s, got=%s", tt.expected, str.Value)
		}
	}
}

func TestCryptoHashErrors(t *testing.T) {
	tests := []string{
		`crypto.sha256()`,    // wrong args
		`crypto.sha256(123)`, // wrong type
		`crypto.sha512()`,    // wrong args
		`crypto.sha512(123)`, // wrong type
	}

	for _, tt := range tests {
		evaluated := testEval(tt)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Errorf("expected error for %s, got=%T", tt, evaluated)
		}
	}
}
