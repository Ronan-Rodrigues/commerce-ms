package security_test

import (
	"testing"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/infra/security"
)

func TestArgon2Hasher_HashAndCompare(t *testing.T) {
	// Parâmetros mais leves para os testes rodarem rápido
	params := security.Argon2Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hasher := security.NewArgon2Hasher(params)
	password := "minhaSenhaUltraSegura#2026"

	// 1. Gera o hash
	hash1, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("falha ao gerar hash da senha: %v", err)
	}

	// 2. Gera segundo hash da mesma senha
	hash2, err := hasher.HashPassword(password)
	if err != nil {
		t.Fatalf("falha ao gerar segundo hash: %v", err)
	}

	// 3. Os hashes DEVEM ser diferentes por causa dos salts aleatórios
	if hash1 == hash2 {
		t.Errorf("dois hashes da mesma senha não devem ser iguais (salt deve variar)")
	}

	// 4. Comparação com a senha correta deve dar true
	if !hasher.ComparePassword(password, hash1) {
		t.Errorf("esperava validação com sucesso para a senha correta no hash1")
	}

	if !hasher.ComparePassword(password, hash2) {
		t.Errorf("esperava validação com sucesso para a senha correta no hash2")
	}

	// 5. Comparação com senha errada deve dar false
	if hasher.ComparePassword("senhaErrada", hash1) {
		t.Errorf("senha incorreta não deveria ser aceita")
	}

	// 6. Hash corrompido deve retornar false sem causar panic
	if hasher.ComparePassword(password, "hash_invalido_qualquer") {
		t.Errorf("hash malformatado não deve ser aceito")
	}
}
