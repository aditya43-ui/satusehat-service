# 🛠️ GoPrint Code Generator Tools

Kumpulan skrip otomatisasi (*Code Generator*) ini dirancang untuk mempercepat proses *development* aplikasi di ekosistem GoPrint. 
Dengan skrip ini, Anda dapat men-generate *boilerplate* (kode dasar) ratusan baris untuk arsitektur **Clean Architecture**, **CQRS**, **REST API**, dan **gRPC** secara instan.

Generator ini sangat cerdas, ia dapat membangun kode berdasarkan **Skema Tabel Database (SQL)** maupun **Respons Payload (JSON)** dari API pihak ketiga.

---

## 1. 🚀 Advanced Context Generator (`context.sh`)

Skrip utama (*Swiss Army Knife*) untuk men-generate seluruh lapisan modul aplikasi Anda.

**Fitur Unggulan:**
- 🧩 **Dual Parser**: Mendukung input dari file SQL (`CREATE TABLE`) maupun file JSON (`Response API`).
- 🔑 **Dynamic Primary Key**: Otomatis mendeteksi *Primary Key* tipe `int64` maupun `string/UUID` dan menyesuaikan *routing* ID-nya.
- 🗑️ **Smart Soft-Delete**: Hanya membuat implementasi GORM *Soft-Delete* jika tabel/JSON tersebut benar-benar memiliki field `deleted_at`.
- 📁 **Auto-Package Naming**: Anda bebas meletakkan modul di *path* mana pun, nama *package* Go akan otomatis menyesuaikan nama folder terakhir.
- 📡 **All-in-One Generation**: Sekali klik langsung membuat: `Entity`, `DTO`, `Mapper`, `Repository (CQRS)`, `Service`, `REST Handler`, `Proto File`, dan `gRPC Handler`.

### 📝 Aturan Penulisan (CLI Syntax):
```bash
./scripts/context.sh [OPTIONS]

Options:
  -s, --sql FILE         [Wajib*] Path ke file SQL yang berisi CREATE TABLE (Pilih salah satu dengan -j)
  -j, --json FILE        [Wajib*] Path ke file JSON payload response (Pilih salah satu dengan -s)
  -d, --dir PATH         [Wajib] Path destinasi modul/folder Anda (contoh: master/role)
  -t, --table NAME       [Opsional] Menimpa nama tabel secara manual (Sangat disarankan saat memakai -j)
  -g, --generate TYPE    [Opsional] Target generate: domain | handler | proto | grpc | all (Default: all)
  -v, --verbose          [Opsional] Tampilkan log proses secara mendetail
# Contoh menggunakan file users.sql dan disimpan di master/users
./scripts/context.sh -s db/migrations/users.sql -d master/users -g all

# Contoh riil menggunakan file role_pages.sql dan disimpan di master/role/pages
./scripts/context.sh -s internal/infrastructure/database/sql/00001_role_pages.sql -d master/role/pages -g all
# Contoh menggunakan file users.sql dan disimpan di master/users
./scripts/context.sh -s db/migrations/users.sql -d master/users -g all
./scripts/context.sh -s internal/infrastructure/database/sql/province.sql -d master/reference/province -g grpc
./scripts/context.sh -s db/migrations/users.sql -d master/users -v
internal/
└── master/
    └── reference/
        └── province/
            ├── dto.go                 # Struct Request & Response (Auto-validation tags)
            ├── entity.go              # GORM Struct (Auto DB tags)
            ├── repository.go          # CQRS Repository (Command & Query builder)
            ├── service.go             # Business Logic Layer
            ├── service_test.go        # Boilerplate Unit Test & Mocking
            └── mapper.go              # Logic konversi Entity <-> DTO

internal/
└── infrastructure/
    └── transport/
        ├── http/
        │   └── handlers/master/reference/province/
        │       └── province_handler.go # REST API Controller (Gin) dgn anotasi Swagger
        └── grpc/ 
            ├── handlers/master/reference/province/
            │   ├── province_grpc_handler.go
            │   └── province_grpc_mapper.go
            ├── proto/master/reference/province/v1/
            │   └── province.proto      # Schema antarmuka gRPC
            └── gen/master/reference/province/v1/
                └── ... (Hasil Compile .pb.go)
./scripts/proto.sh
./scripts/proto.sh internal/infrastructure/transport/grpc/proto/master/reference/province/v1
# 1. Mengecek daftar service yang terbuka (tes koneksi Server Reflection)
grpcurl -plaintext localhost:50051 list

# 2. Mengecek rincian method yang dimiliki oleh suatu service
grpcurl -plaintext localhost:50051 list master.v1.RoleAccessRolMasterService

# 3. Memanggil Method dengan mengirim Payload JSON
grpcurl -plaintext -d '{
  "id": 1
}' localhost:50051 master.v1.RoleAccessRolMasterService/GetRoleAccessRolMaster

<!--
[PROMPT_SUGGESTION]Bagaimana cara kerja fitur `-j` atau `--json` pada `context.sh` untuk mem-parsing payload API eksternal?[/PROMPT_SUGGESTION]
[PROMPT_SUGGESTION]Jelaskan alur registrasi handler REST dan gRPC di `main.go` setelah sebuah modul baru di-generate.[/PROMPT_SUGGESTION]
-->


./scripts/grpc.sh -d master/reference/province  