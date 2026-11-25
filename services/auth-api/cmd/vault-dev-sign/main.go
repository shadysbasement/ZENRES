// Command vault-dev-sign is a DEVELOPMENT stand-in for `gpg --decrypt | gpg --clearsign` so the hidden
// vault can be exercised end-to-end on machines without GnuPG. Never use it with production keys.
//
//	go run ./auth-api/cmd/vault-dev-sign keygen -out dev-admin      # writes dev-admin.asc + dev-admin.key
//	go run ./auth-api/cmd/vault-dev-sign sign -key dev-admin.key < challenge.asc
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vault-dev-sign keygen|sign …")
		os.Exit(2)
	}
	pgp := crypto.PGPWithProfile(profile.RFC9580())
	switch os.Args[1] {
	case "keygen":
		fs := flag.NewFlagSet("keygen", flag.ExitOnError)
		out := fs.String("out", "dev-admin", "output basename")
		_ = fs.Parse(os.Args[2:])
		k, err := pgp.KeyGeneration().AddUserId("Zevra Dev Admin", "dev-admin@zevra.local").New().GenerateKey()
		must(err)
		pub, _ := k.GetArmoredPublicKey()
		priv, _ := k.Armor()
		must(os.WriteFile(*out+".asc", []byte(pub), 0o644))
		must(os.WriteFile(*out+".key", []byte(priv), 0o600))
		fmt.Printf("fingerprint %X\nwrote %s.asc (public) and %s.key (PRIVATE — dev only)\n", k.GetFingerprintBytes(), *out, *out)
	case "sign":
		fs := flag.NewFlagSet("sign", flag.ExitOnError)
		keyPath := fs.String("key", "dev-admin.key", "armored private key")
		_ = fs.Parse(os.Args[2:])
		raw, err := os.ReadFile(*keyPath)
		must(err)
		k, err := crypto.NewKeyFromArmored(string(raw))
		must(err)
		challenge, _ := io.ReadAll(os.Stdin)
		dec, err := pgp.Decryption().DecryptionKey(k).New()
		must(err)
		res, err := dec.Decrypt(challenge, crypto.Armor)
		must(err)
		s, err := pgp.Sign().SigningKey(k).New()
		must(err)
		out, err := s.SignCleartext(res.Bytes())
		must(err)
		_, _ = os.Stdout.Write(out)
	default:
		fmt.Fprintln(os.Stderr, "usage: vault-dev-sign keygen|sign …")
		os.Exit(2)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "vault-dev-sign:", err)
		os.Exit(1)
	}
}
