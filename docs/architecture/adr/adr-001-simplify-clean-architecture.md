# ADR-001: Sederhanakan Clean Architecture (buang ceremony, pertahankan prinsip)

- **Status:** Accepted (sebagian diimplementasikan — lihat Status implementasi)
- **Tanggal:** 2026-06-18
- **Cakupan:** Penyederhanaan struktur internal. **Tidak** mencakup roadmap microservice.

## Konteks

Backend mengikuti Clean Architecture. *Dependency rule* (adapter → usecase →
domain) sudah dipatuhi dengan benar. Namun struktur terasa *bloated* untuk domain
yang pada dasarnya CRUD (workspace → board → column → card):

1. **DTO seremonial.** Tiap entity punya struct `dto.*` yang menyalin field
   entity 1:1 tanpa transformasi, plus 2–3 fungsi mapping per entity.
2. **Interface use case tanpa konsumen yang butuh substitusi.** Tiap use case
   punya interface + `…Impl`, padahal hanya ada satu implementasi dan **nol unit
   test** di seluruh `internal/usecase`.
3. **Penamaan `…Impl`** yang tidak idiomatik di Go.

Biaya abstraksi dibayar; manfaatnya (testability, multiple implementations) belum
dipanen.

## Keputusan

### 1. Buang lapisan `internal/dto`
Use case mengembalikan domain entity langsung (dibungkus struct hasil kecil bila
perlu, mis. `CardResult{Card *entity.Card; Assignee *entity.User}`). Pemetaan ke
JSON terjadi di `delivery/http/response`, membaca dari entity.

**Pengecualian:** DTO boleh dipertahankan **jika** output use case butuh bentuk
yang benar-benar beda dari entity (gabungan multi-entity, sembunyikan field
sensitif). Nilai per kasus; jangan menyalin entity 1:1.

### 2. Use case: return struct konkret, bukan interface
Hapus interface `…UseCase`. Constructor return `*…UseCase`. Handler bergantung
pada tipe konkret (`*card.CardUseCase`).

**Jika** nanti butuh mock untuk test handler: definisikan interface kecil **di
package handler** (sisi konsumen), berisi hanya method yang dipakai.

### 3. Repository & `BoardAccessChecker`: TETAP interface
Keduanya punya alasan sah (mock untuk test + kemungkinan ganti implementasi;
`BoardAccessChecker` dipakai 3 modul). Interface dipertahankan. Hanya penamaan
yang dirapikan (lihat #4).

### 4. Buang suffix `…Impl`
Tipe konkret diberi nama deskriptif. Untuk komponen yang interface-nya
dipertahankan, jadikan tipe konkret **unexported** (`type cardRepository struct`)
karena yang diekspor hanya constructor.

## Prinsip pemandu

> **Accept interfaces, return structs.** Definisikan interface di sisi konsumen,
> sekecil mungkin, hanya saat ada konsumen yang butuh menukar implementasi
> (test mock atau varian nyata). Lapisan data hanya dibuat jika ia
> *mentransformasi*, bukan sekadar memindahkan.

Tabel keputusan:

| Komponen | Interface dipertahankan? | Alasan |
|---|---|---|
| `*Repository` | Ya | Mock test + bisa ganti backend; interface di sisi konsumen (`domain`) |
| `common.BoardAccessChecker` | Ya | 3 konsumen lintas-modul |
| `*UseCase` | Tidak | 1 implementasi, 0 test; tak ada konsumen yang butuh substitusi |

## Konsekuensi

**Positif**
- Lebih sedikit kode boilerplate; alur data lebih mudah diikuti.
- Lebih idiomatik Go; ramah pendatang baru.
- Batas modul lebih jelas — fondasi yang lebih baik untuk pemecahan service kelak.

**Negatif / risiko**
- Handler kini terikat tipe use case konkret. Mitigasi: interface mock
  diperkenalkan di sisi handler saat test ditulis.
- Refactor menjalar antar modul (mis. jalur kanban). Mitigasi: dikerjakan
  per modul dengan jembatan sementara ber-`TODO`; build dijaga tetap hijau.

## Status implementasi

**Selesai seluruhnya** — `go build ./...`, `go vet ./...`, `gofmt -l internal/`,
dan `go test ./...` semua bersih/lulus.

- ✅ Modul `card` — selesai & tervalidasi.
- ✅ Modul `column` — selesai & tervalidasi.
- ✅ Modul `board` — selesai & tervalidasi (DTO transformatif dipindah ke package).
- ✅ Modul `workspace` — selesai & tervalidasi.
- ✅ Modul `auth` — selesai & tervalidasi (`UserDTO` → `auth.UserProfile`,
  dipertahankan karena menyembunyikan `PasswordHash`).
- ✅ **`internal/dto` dihapus seluruhnya** — tak ada lagi DTO global.
- ✅ Update `docs/architecture.md` — tabel "Data layers", folder map, dan langkah
  endpoint sudah disesuaikan.

### Tidak dikerjakan (sengaja, di luar scope ADR ini)

- Unit test usecase belum ditulis. Ini *manfaat* yang diaktifkan refactor
  (testability lebih bersih), tapi bukan bagian dari penyederhanaan itu sendiri.
  Saat ditulis nanti, interface mock diperkenalkan **di sisi handler**, bukan
  menghidupkan kembali interface use case.

Detail langkah per modul: lihat [refactor-playbook.md](./refactor-playbook.md).
Penjelasan mendalam + contoh: lihat [refactor-learning-notes.md](./refactor-learning-notes.md).
