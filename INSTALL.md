# Installation

🇬🇧 [English](#english) | 🇮🇩 [Indonesia](#indonesia)

---

<a name="english"></a>
## 🇬🇧 English

### Install

**Download** the `forgecode` binary for your platform, then:

```bash
chmod +x forgecode       # Linux/Mac only
./forgecode --install    # Install
```

### Update

To update, simply download the latest binary and run the install command again:

```bash
./forgecode --install    # This will overwrite the old version
```

Restart terminal, then:

```bash
forgecode --version      # Verify
```

### Uninstall

```bash
forgecode --uninstall
```

### Troubleshooting

**"forgecode not found"** - Restart your terminal

**Linux/Mac PATH issue:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

<a name="indonesia"></a>
## 🇮🇩 Indonesia

### Install

**Download** binary `forgecode` untuk platform lu, terus:

```bash
chmod +x forgecode       # Khusus Linux/Mac
./forgecode --install    # Install
```

### Update

Buat update, tinggal download binary terbaru terus jalanin lagi command install-nya:

```bash
./forgecode --install    # Ini bakal numpuk (overwrite) versi lama
```

Restart terminal, abis itu:

```bash
forgecode --version      # Cek versi
```

### Uninstall

```bash
forgecode --uninstall
```

### Troubleshooting

**"forgecode not found"** - Restart terminal lu

**Linux/Mac masalah PATH:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

**License:** MIT
