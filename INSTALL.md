# Installation

🇬🇧 [English](#english) | 🇮🇩 [Indonesia](#indonesia)

---

<a name="english"></a>
## 🇬🇧 English

### Install

**Download** the `curecode` binary for your platform, then:

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

### Install

**Download** binary `curecode` untuk platform lu, terus:

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
