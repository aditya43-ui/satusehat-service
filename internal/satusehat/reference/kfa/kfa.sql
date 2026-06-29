-- Tabel Utama Produk KFA
CREATE TABLE kfa_products (
    kfa_code VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    nama_dagang VARCHAR(255),
    active BOOLEAN DEFAULT true,
    state VARCHAR(50),
    generik BOOLEAN,
    
    -- Klasifikasi & Atribut Dasar (Disimpan sebagai JSONB untuk efisiensi)
    dosage_form JSONB,         -- { "code": "BS066", "name": "Tablet" }
    farmalkes_type JSONB,      -- { "code": "medicine", "group": "farmasi", "name": "Obat" }
    uom JSONB,                 -- { "name": "Tablet" }
    rute_pemberian JSONB,      -- { "code": "O", "name": "Oral" }
    controlled_drug JSONB,     -- { "code": "3", "name": "Obat Keras" }
    
    -- Harga & Metrik Fisik
    fix_price NUMERIC(12, 2),
    het_price NUMERIC(12, 2),
    dose_per_unit NUMERIC(10, 2),
    net_weight NUMERIC(10, 2),
    net_weight_uom_name VARCHAR(20),
    volume NUMERIC(10, 2),
    volume_uom_name VARCHAR(20),
    score_tkdn NUMERIC(5, 2),
    rxterm INT,
    
    -- Data Legal & Manufaktur
    manufacturer VARCHAR(255),
    registrar VARCHAR(255),
    nie VARCHAR(100),
    farmalkes_hscode VARCHAR(100),
    kode_lkpp VARCHAR(100),
    tayang_lkpp BOOLEAN,
    
    -- Teks Panjang (HTML String)
    description TEXT,
    indication TEXT,
    side_effect TEXT,
    warning TEXT,
    
    -- Kompleks/Nested JSONB (Data yang jarang di-join tapi sering dibaca)
    atc_info JSONB,            -- Menggabungkan atc_l1 sampai atc_l5 & atc_ddd
    fornas JSONB,              -- Struktur dokumen fornas
    dosage_usage JSONB,        -- Array of dosage usage objects
    tags JSONB,                -- Array of tags [{code, name}]
    identifier_ids JSONB,      -- Array of identifiers [{code, name, source_name}]
    product_template JSONB,    -- Data template referensi KFA
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indeks untuk pencarian teks yang cepat
CREATE INDEX idx_kfa_products_name ON kfa_products(name);
CREATE INDEX idx_kfa_products_kode_lkpp ON kfa_products(kode_lkpp);
-- Indeks GIN untuk pencarian dalam JSONB (misal mencari produk berdasarkan tags)
-- CREATE INDEX idx_kfa_products_tags ON kfa_products USING GIN (tags); -- Catatan: Buka komentar ini HANYA jika Anda menggunakan PostgreSQL

-- Tabel Relasi: Zat Aktif (Penting untuk dibuat tabel terpisah agar bisa query "Obat dengan kandungan X")
CREATE TABLE kfa_active_ingredients (
    id SERIAL PRIMARY KEY,
    product_kfa_code VARCHAR(50) REFERENCES kfa_products(kfa_code) ON DELETE CASCADE,
    kfa_code VARCHAR(50),
    zat_aktif VARCHAR(255) NOT NULL,
    kekuatan_zat_aktif VARCHAR(100),
    active BOOLEAN DEFAULT true,
    state VARCHAR(50),
    updated_at TIMESTAMP
);
CREATE INDEX idx_ingredient_zat_aktif ON kfa_active_ingredients(zat_aktif);

-- Tabel Relasi: Varian Kemasan
CREATE TABLE kfa_packaging (
    id SERIAL PRIMARY KEY,
    product_kfa_code VARCHAR(50) REFERENCES kfa_products(kfa_code) ON DELETE CASCADE,
    kfa_code VARCHAR(50),
    name VARCHAR(255),
    pack_price NUMERIC(12, 2),
    qty INT,
    uom_id VARCHAR(50)
);