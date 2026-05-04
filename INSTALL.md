# Installation

🇬🇧 [English](#english) | 🇮🇩 [Indonesia](#indonesia)

---

<a name="english"></a>
## 🇬🇧 English

**Option 1: One-liner (Recommended)**
```bash
curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash
```

**Option 2: Manual Download**
Download the `curecode` binary for your platform, then:
```bash
chmod +x curecode       # Linux/Mac only
./curecode --install    # Install
```

### Update

To update, simply download the latest binary and run the install command again:

```bash
./curecode --install    # This will overwrite the old version
```

Restart terminal, then:

```bash
curecode --version      # Verify
```

### Uninstall

```bash
curecode --uninstall
```

### Troubleshooting

**"curecode not found"** - Restart your terminal

**Linux/Mac PATH issue:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

<a name="indonesia"></a>
## 🇮🇩 Indonesia

**Opsi 1: One-liner (Rekomendasi)**
```bash
curl -fsSL https://raw.githubusercontent.com/broman0x/cure-code/main/install.sh | bash
```

**Opsi 2: Manual Download**
Download binary `curecode` untuk platform lu, terus:
```bash
chmod +x curecode       # Khusus Linux/Mac
./curecode --install    # Install
```

### Update

Buat update, tinggal download binary terbaru terus jalanin lagi command install-nya:

```bash
./curecode --install    # Ini bakal numpuk (overwrite) versi lama
```

Restart terminal, abis itu:

```bash
curecode --version      # Cek versi
```

### Uninstall

```bash
curecode --uninstall
```

### Troubleshooting

**"curecode not found"** - Restart terminal lu

**Linux/Mac masalah PATH:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

**License:** MIT
