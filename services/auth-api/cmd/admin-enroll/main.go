// Command admin-enroll registers a founding-team OpenPGP PUBLIC key into admin_pgp_keys.
// It is the only way to create a vault identity — there is no web signup for admins.
//
//	gpg --armor --export 4A9F...88C1 > founder.asc
//	DATABASE_URL=postgres://... go run ./auth-api/cmd/admin-enroll -handle zv-founder-01 -key founder.asc
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shadysbasement/zevra/services/auth-api/internal/pgp"
	"github.com/shadysbasement/zevra/services/auth-api/internal/store"
)

func main() {
	handle := flag.String("handle", "", "admin handle, e.g. zv-founder-01")
	keyPath := flag.String("key", "", "path to ASCII-armored PUBLIC key (.asc)")
	role := flag.String("role", "super_admin", "admin role")
	dry := flag.Bool("dry-run", false, "print the SQL instead of writing to the database")
	flag.Parse()
	if *handle == "" || *keyPath == "" {
		flag.Usage()
		os.Exit(2)
	}
	raw, err := os.ReadFile(*keyPath)
	must(err)
	fpr, key, err := pgp.FingerprintOf(string(raw))
	must(err)
	if !key.CanVerify(time.Now().Unix()) {
		must(fmt.Errorf("key cannot produce valid signatures (expired or revoked)"))
	}
	pub, err := key.GetArmoredPublicKey()
	must(err)
	fmt.Printf("Handle:      %s\nFingerprint: %s\nRole:        %s\n", *handle, groups(fpr), *role)
	if *dry {
		fmt.Printf("\nINSERT INTO admin_pgp_keys (handle, fingerprint, armored_public_key, role) VALUES ('%s', '%s', $$%s$$, '%s');\n", *handle, fpr, pub, *role)
		return
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		must(fmt.Errorf("DATABASE_URL is required (or use -dry-run)"))
	}
	pg, err := store.NewPostgres(context.Background(), dsn)
	must(err)
	must(pg.InsertAdminKey(context.Background(), &store.AdminKey{Handle: *handle, Fingerprint: fpr, ArmoredPublicKey: pub, Role: *role, Status: "active"}))
	fmt.Println("\n✓ Enrolled. The private key never left your machine.")
}

func groups(f string) string {
	var b strings.Builder
	for i, c := range f {
		if i > 0 && i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "admin-enroll:", err)
		os.Exit(1)
	}
}
