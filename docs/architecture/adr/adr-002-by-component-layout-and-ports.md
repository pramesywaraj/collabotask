# ADR-002: Layout by-component, dependency inversion auth, & validator ke pkg

- **Status:** Accepted & diimplementasikan
- **Tanggal:** 2026-06-22
- **Konteks lanjutan dari:** [ADR-001](./adr-001-simplify-clean-architecture.md)

Tiga keputusan terkait yang diambil dalam satu sesi. Semuanya memperkuat
*dependency rule* Clean Architecture; tidak ada yang mengubah logika bisnis.

---

## Keputusan 1 — Layout folder: by-component (bukan `adapter/` umbrella)

### Konteks
Sebelumnya transport & persistence dikelompokkan di bawah satu `adapter/`
(`adapter/http`, `adapter/repository/postgres`) — gaya "by-layer" / hexagonal.
Setelah membandingkan dengan project lain di organisasi (`quota-worker`,
`ops-supply-chain`) yang konsisten memakai layout **by-component** (folder
top-level per jenis komponen), diputuskan menyelaraskan collabotask.

### Keputusan
- `adapter/http/` → `delivery/http/`
- `adapter/repository/postgres/` → `repository/postgres/`
- `adapter/` dihapus sepenuhnya.

**Invariant yang dijaga ketat:** interface repository **tetap** di
`domain/repository/` (sisi konsumen). Hanya *implementasi* yang pindah ke
`repository/postgres/`. Implementasi bergantung pada interface domain — tidak
sebaliknya. Inilah yang menjaga dependency rule tetap utuh meski folder pindah.

### Alasan
- **Konsistensi lintas-project** di organisasi (DX bagi developer yang berpindah
  antar-repo) — ini alasan utama, bukan kemurnian teoretis.
- by-component dan by-layer sama-sama sah secara Clean Architecture; yang
  menentukan benar/salah adalah *arah dependency*, bukan nama folder.

### Catatan
Keputusan ini **murni penamaan/penempatan folder** — 14 file berubah, hanya
import path. Nol perubahan logika. Migrasi setengah jalan (mencampur `adapter/`
dengan by-component) dihindari secara sadar; vocabulary harus konsisten penuh.

---

## Keputusan 2 — Dependency inversion untuk auth (ports)

### Konteks
`usecase/auth` memanggil **fungsi konkret** `infrastructure/auth`
(`HashPassword`, `GenerateToken`, …) secara langsung. Ini melanggar dependency
rule: use case (lingkaran dalam) tahu detail bcrypt & JWT (lingkaran luar). Efek
samping: use case auth tidak bisa di-unit-test tanpa menjalankan bcrypt/JWT asli.

### Keputusan
Definisikan **port** di sisi konsumen (`usecase/auth/auth_usecase.go`):

```go
type PasswordHasher interface { Hash(...) ...; Check(...) bool }
type TokenGenerator interface { Generate(...) (string, error) }
```

`infrastructure/auth` menyediakan implementasinya (`BcryptHasher`,
`JWTGenerator`) yang menyerap `*config.AuthConfig`. Wiring di `injection`.

### Konsekuensi
- `usecase/auth` kini **nol import** `infrastructure` — algoritma hashing/token
  bisa diganti (argon2, PASETO) tanpa menyentuh use case.
- `config` ikut lepas dari use case auth.
- Use case auth siap di-unit-test dengan mock port (lihat ADR-001: testability
  yang belum dipanen).
- Compile-time assertion (`var _ auth.PasswordHasher = (*BcryptHasher)(nil)`) di
  `infrastructure/auth/provider.go` menjaga adapter selalu memenuhi port.

### Batas
Pola port ini **hanya** untuk dependency yang punya I/O / secret / bisa di-swap
(auth). **Tidak** diterapkan ke utility stateless deterministik — lihat
Keputusan 3.

---

## Keputusan 3 — `validator` pindah dari `infrastructure/` ke `pkg/`

### Konteks
`validator` (wrapper `go-playground/validator`) berada di
`infrastructure/validator`, sehingga 23 use case "import infrastructure" —
membuat audit dependency-rule (`grep usecase → infrastructure`) menyala merah
sebagai false alarm.

### Keputusan
Pindah ke `pkg/validator`. Package tidak berubah isinya.

### Alasan
- `validator` adalah **stateless utility deterministik** (nol I/O, nol secret) —
  bukan "infrastruktur" dalam arti menyembunyikan detail eksternal. Ia setara
  `fmt`/`strings`/`time`, sekategori dengan `pkg/logger`.
- **Bukan** kasus untuk port/interface (beda dari auth) — solusinya *pindah
  package*, bukan *menambah abstraksi*. Memberi validator interface = abstraksi
  tanpa alasan.
- Efek: bersama Keputusan 2, `usecase` kini **nol import `infrastructure`** →
  dependency-rule grep bersih sempurna, sinyal audit jadi jujur.

---

## Status keseluruhan

Semua tervalidasi: `go build ./...`, `go vet ./...`, `gofmt -l`, `go test ./...`
bersih/lulus. `usecase` & `domain` tidak meng-import HTTP/DB/framework.

Dokumen terkait diselaraskan: `README.md`, `docs/architecture.md` (layout +
folder map), dan referensi usang di ADR-001 & learning-notes.
