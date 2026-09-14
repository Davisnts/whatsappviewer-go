# WhatsApp Viewer (Go Reimplementation)

> 🚀 **A modern, cross-platform WhatsApp database viewer, decrypter, and exporter written in pure Go — supporting modern (2020–2026+) and legacy Android databases (`msgstore.db`), featuring an embedded WhatsApp Web-style UI.**
> 
> *Um visualizador, descriptografador e exportador moderno e multiplataforma de bancos de dados do WhatsApp desenvolvido em Go puro — compatível com versões recentes (2020–2026+) e legadas do WhatsApp Android (`msgstore.db`), com interface embutida estilo WhatsApp Web.*

[![Go Report Card](https://goreportcard.com/badge/github.com/andreas-mausch/whatsapp-viewer)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-brightgreen.svg)]()
[![Pure Go](https://img.shields.io/badge/Pure%20Go-No%20CGO-00ADD8.svg?logo=go)]()

> 🌐 **Language / Idioma:** [🇬🇧 English](#-english) | [🇧🇷 Português](#-português)

---

<a name="-english"></a>
## 🇬🇧 English

### Overview

**WhatsApp Viewer (Go)** is a modern, high-performance reimplementation in **Go (Golang)** of the original [WhatsApp Viewer by Andreas Mausch](https://github.com/andreas-mausch/whatsapp-viewer).

While the original C++ tool was revolutionary, newer WhatsApp database updates (2020–2026+) introduced structural schema alterations (such as modern table layouts `message`, `chat`, `jid`, complex views, new index formats like `lid_display_name_upper_username_index`, and AES-GCM encryption requirements) that often cause legacy SQLite drivers and C++ versions to crash or fail.

This rewrite completely modernizes the architecture:
- **100% Pure Go** (no CGO required, thanks to `modernc.org/sqlite`).
- **Fully compatible with the latest WhatsApp database formats** while maintaining backward compatibility with legacy formats.
- **Cross-platform**: Runs effortlessly on Windows, Linux, and macOS.
- **Modern Web Interface**: Built-in, responsive WhatsApp Web-style interface embedded in the binary.
- **Rich Export Engine**: Export conversations to HTML, JSON, or Plain Text.

---

### Key Features

* ⚡ **Full Compatibility with Modern WhatsApp (2020–2026+)**: Seamlessly handles modern tables (`chat`, `jid`, `message`, `message_media`, `message_quoted`, etc.) and legacy tables (`messages`, `chat_view`) with automatic schema detection.
* 🛡️ **No Index Crash Issues**: Immune to SQLite index conflicts such as `lid_display_name_upper_username_index` without needing manual index dropping.
* 🌐 **Embedded WhatsApp Web UI**: Fast local HTTP server serving a single-page app mimicking WhatsApp Web with dark/light themes, search, media previews, thumbnails, quote replies, and geo-locations.
* 🔐 **Crypt Decryption Support**: Decrypts Android WhatsApp database backups (`.crypt5`, `.crypt7`, `.crypt8`, `.crypt12`, and `.crypt14`) directly via CLI.
* 📦 **Multi-Format Chat Export**:
  * **HTML**: Beautiful, responsive, standalone transcript ready for sharing or printing.
  * **JSON**: Complete structured data for developers, archival, and data analysis.
  * **TXT**: Clean, chronological plain-text transcript.
* 📇 **Contacts Integration**: Optionally load `wa.db` to display actual contact names instead of phone numbers/JIDs.
* 🚀 **Single Standalone Binary**: Zero runtime dependencies. No DLLs or runtimes to install.

---

### Comparison: Original C++ vs. Go Reimplementation

| Feature | Original WhatsApp Viewer (C++) | Go Reimplementation |
| :--- | :--- | :--- |
| **Language** | C++ (MSVC / MFC / WTL) | Go 1.20+ |
| **Platform** | Windows only | Windows, Linux, macOS |
| **Modern WhatsApp (2020+)** | ⚠️ Incompatible / requires manual SQL fixes | ✅ Native automatic support |
| **Legacy WhatsApp** | ✅ Supported | ✅ Supported (Auto-detected) |
| **Database Engine** | C++ SQLite library (older) | Pure Go SQLite (`modernc.org/sqlite`) - No CGO |
| **User Interface** | Win32 MFC Desktop GUI | Modern WhatsApp Web-style Embedded SPA |
| **Export Formats** | HTML, TXT, JSON | HTML (modern layout), JSON, TXT |
| **Decryption** | crypt5, 7, 8, 12, 14 | crypt5, 7, 8, 12, 14 (AES-GCM & Zlib) |
| **Portability** | Requires VC++ runtimes | Single static executable |

---

### How to Get Your WhatsApp Database

To view WhatsApp chats on your computer, you need the SQLite database and (optionally) key and contact files extracted from your Android device:

> [!NOTE]
> Accessing these files on Android typically requires **root access**, an unlocked bootloader, an emulator, or an Android backup extraction tool.

| File Path on Android | Purpose |
| :--- | :--- |
| `/data/data/com.whatsapp/databases/msgstore.db` | Main decrypted chat database |
| `/data/data/com.whatsapp/databases/wa.db` | Contacts database (optional, for contact names) |
| `/data/data/com.whatsapp/files/key` | AES decryption key file (for `.crypt` backups) |
| `/sdcard/WhatsApp/Databases/msgstore.db.crypt14` | Encrypted backup file |

---

### Usage

#### 1. Launching the Web UI

Simply run the executable passing your `msgstore.db` (and optionally `wa.db`):

```bash
# Direct launch (auto-opens your default browser at http://localhost:8080)
whatsapp-viewer.exe msgstore.db

# Or using the serve subcommand
whatsapp-viewer.exe serve -db msgstore.db -wa wa.db -port 8080
```

#### 2. Decrypting Backups

Decrypt Android encrypted backups using your `key` file:

```bash
# Crypt14 (Most recent)
whatsapp-viewer.exe -decrypt14 msgstore.db.crypt14 key msgstore.decrypted.db

# Crypt12
whatsapp-viewer.exe -decrypt12 msgstore.db.crypt12 key msgstore.decrypted.db

# Crypt8
whatsapp-viewer.exe -decrypt8 msgstore.db.crypt8 key msgstore.decrypted.db

# Crypt7
whatsapp-viewer.exe -decrypt7 msgstore.db.crypt7 key msgstore.decrypted.db

# Crypt5 (Uses account email instead of key file)
whatsapp-viewer.exe -decrypt5 msgstore.db.crypt5 your-account@gmail.com msgstore.decrypted.db
```

#### 3. Exporting Chats

Export conversations to HTML, JSON, or Plain Text directly from the terminal:

```bash
# Bulk export all chats to HTML in a folder named 'exports'
whatsapp-viewer.exe export -db msgstore.db -wa wa.db -format html -out ./exports

# Export a single chat by JID to JSON
whatsapp-viewer.exe export -db msgstore.db -jid "1234567890@s.whatsapp.net" -format json -out chat.json

# Export all chats to plain text (.txt)
whatsapp-viewer.exe export -db msgstore.db -format txt -out ./txt_exports
```

---

### Building from Source

Ensure you have **Go 1.20+** installed:

```bash
# Clone the repository
git clone https://github.com/andreas-mausch/whatsapp-viewer.git
cd "whatsapp-viewer/go refact"

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary for current platform
go build -o whatsapp-viewer.exe .
```

#### Cross-compilation:
```bash
# Build for Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o whatsapp-viewer-linux .

# Build for macOS (Apple Silicon M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o whatsapp-viewer-darwin-arm64 .

# Build for Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o whatsapp-viewer.exe .
```

---

### Credits & Acknowledgments

* **Andreas Mausch** ([@andreas-mausch](https://github.com/andreas-mausch)): Creator of the original [whatsapp-viewer](https://github.com/andreas-mausch/whatsapp-viewer) in C++.
* **TripCode & EliteAndroidApps**: Crypt12 reverse engineering insights.
* **@torsade & contributors**: Crypt14 structure contributions.
* **modernc.org/sqlite**: Pure Go SQLite implementation enabling CGO-free portability.

---

<br/>

---

<a name="-português"></a>
## 🇧🇷 Português

### Visão Geral

O **WhatsApp Viewer (Go)** é uma reimplementação moderna e de alta performance em **Go (Golang)** do projeto original [WhatsApp Viewer de Andreas Mausch](https://github.com/andreas-mausch/whatsapp-viewer).

Embora a ferramenta original em C++ tenha sido uma referência pioneira, as atualizações estruturais nas versões recentes do banco de dados do WhatsApp (2020 a 2026+) introduziram modificações substanciais de esquema (tabelas modernas `message`, `chat`, `jid`, separação de mídias em `message_media`, respostas citadas em `message_quoted`, índices como `lid_display_name_upper_username_index` e criptografia AES-GCM) que frequentemente causavam falhas ou exigiam correções manuais no código C++ legado.

Esta reimplementação em Go renova completamente a aplicação:
- **100% Pure Go** (sem necessidade de CGO, utilizando `modernc.org/sqlite`).
- **Suporte nativo às versões mais recentes do WhatsApp**, além de compatibilidade com versões antigas (legadas).
- **Multiplataforma**: Execução nativa no Windows, Linux e macOS sem dependências de bibliotecas C++.
- **Interface Web Moderna**: Interface web integrada embutida no próprio executável, com visual inspirado no WhatsApp Web.
- **Mecanismo de Exportação Flexível**: Exportação direta para HTML, JSON ou TXT.

---

### Funcionalidades Principais

* ⚡ **Compatibilidade com Versões Recentes do WhatsApp (2020–2026+)**: Lê automaticamente tanto esquemas modernos (`chat`, `jid`, `message`, `message_media`, `message_quoted`, etc.) quanto antigos (`messages`, `chat_view`) por detecção dinâmica.
* 🛡️ **Livre de Falhas de Índice do SQLite**: Não trava com novos índices como `lid_display_name_upper_username_index`, dispensando comandos manuais de exclusão de índices.
* 🌐 **Interface Web Embutida (Estilo WhatsApp Web)**: Servidor HTTP local integrado que abre no navegador padrão, oferecendo busca de conversas, fotos/iniciais dos contatos, mensagens citadas, indicação de mídias, miniaturas e visualização de localização geográfica.
* 🔐 **Descriptografia de Backups**: Descriptografa backups do Android (`.crypt5`, `.crypt7`, `.crypt8`, `.crypt12` e `.crypt14`) utilizando a chave (`key`) ou e-mail da conta (crypt5).
* 📦 **Exportação em Múltiplos Formatos**:
  * **HTML**: Arquivo responsivo, moderno e independente, ideal para visualização ou impressão.
  * **JSON**: Estrutura de dados completa para desenvolvedores e análise forense/dados.
  * **TXT**: Registro em texto simples e cronológico.
* 📇 **Suporte a Nomes de Contatos**: Carrega opcionalmente o banco `wa.db` para exibir os nomes reais dos contatos em vez de apenas o número/JID.
* 🚀 **Binário Único e Portátil**: Não requer instalação de interpretadores, servidores web externos ou runtimes do Visual C++.

---

### Comparativo: Versão Original (C++) vs. Reimplementação em Go

| Recurso | Versão Original (C++) | Reimplementação em Go |
| :--- | :--- | :--- |
| **Linguagem** | C++ (MSVC / MFC / WTL) | Go 1.20+ |
| **Plataformas** | Apenas Windows | Windows, Linux, macOS |
| **WhatsApp Moderno (2020+)** | ⚠️ Incompatível / requer correções manuais no SQL | ✅ Suporte nativo automático |
| **WhatsApp Legado** | ✅ Suportado | ✅ Suportado (Detecção automática) |
| **Driver SQLite** | SQLite C++ compilado | Pure Go (`modernc.org/sqlite`) - Sem CGO |
| **Interface de Usuário** | Janela Win32 estilo Windows 95/XP | Interface Web moderna estilo WhatsApp Web |
| **Formatos de Exportação** | HTML, TXT, JSON | HTML (moderno responsivo), JSON, TXT |
| **Descriptografia** | crypt5, 7, 8, 12, 14 | crypt5, 7, 8, 12, 14 (AES-GCM e Zlib) |
| **Portabilidade** | Depende de runtimes do VC++ | Binário estático único sem dependências |

---

### Como Obter os Arquivos do WhatsApp no Android

Para visualizar suas conversas no computador, você precisa do banco de dados SQLite e, opcionalmente, dos arquivos de chave e contatos extraídos do seu aparelho Android:

> [!NOTE]
> O acesso direto a esses arquivos no Android normalmente exige **acesso root**, bootloader desbloqueado, uso de emulador ou ferramentas de extração de backup.

| Caminho no Android | Descrição |
| :--- | :--- |
| `/data/data/com.whatsapp/databases/msgstore.db` | Banco de dados principal das mensagens (descriptografado) |
| `/data/data/com.whatsapp/databases/wa.db` | Banco de dados de contatos (opcional, para nomes) |
| `/data/data/com.whatsapp/files/key` | Arquivo de chave AES (para backups `.crypt`) |
| `/sdcard/WhatsApp/Databases/msgstore.db.crypt14` | Backup criptografado do WhatsApp |

---

### Instruções de Uso

#### 1. Visualização na Interface Web

Abra o executável passando o arquivo `msgstore.db` (e opcionalmente `wa.db`):

```bash
# Execução direta (abre o navegador padrão em http://localhost:8080)
whatsapp-viewer.exe msgstore.db

# Ou através do comando serve
whatsapp-viewer.exe serve -db msgstore.db -wa wa.db -port 8080
```

#### 2. Descriptografia de Backups

Descriptografe backups do Android usando o arquivo `key`:

```bash
# Crypt14 (Versões recentes)
whatsapp-viewer.exe -decrypt14 msgstore.db.crypt14 key msgstore.decrypted.db

# Crypt12
whatsapp-viewer.exe -decrypt12 msgstore.db.crypt12 key msgstore.decrypted.db

# Crypt8
whatsapp-viewer.exe -decrypt8 msgstore.db.crypt8 key msgstore.decrypted.db

# Crypt7
whatsapp-viewer.exe -decrypt7 msgstore.db.crypt7 key msgstore.decrypted.db

# Crypt5 (Usa e-mail da conta em vez do arquivo de chave)
whatsapp-viewer.exe -decrypt5 msgstore.db.crypt5 seu-email@gmail.com msgstore.decrypted.db
```

#### 3. Exportação de Conversas

Exporte conversas em lote ou individualmente via linha de comando:

```bash
# Exportar todas as conversas para HTML na pasta 'exports'
whatsapp-viewer.exe export -db msgstore.db -wa wa.db -format html -out ./exports

# Exportar uma conversa específica por JID para JSON
whatsapp-viewer.exe export -db msgstore.db -jid "5511999999999@s.whatsapp.net" -format json -out chat.json

# Exportar todas as conversas para texto puro (.txt)
whatsapp-viewer.exe export -db msgstore.db -format txt -out ./txt_exports
```

---

### Compilação a partir do Código-Fonte

Certifique-se de ter o **Go 1.20+** instalado:

```bash
# Clonar o repositório
git clone https://github.com/andreas-mausch/whatsapp-viewer.git
cd "whatsapp-viewer/go refact"

# Baixar dependências
go mod download

# Executar testes unitários
go test ./...

# Compilar o executável para a plataforma atual
go build -o whatsapp-viewer.exe .
```

#### Compilação cruzada (Cross-Compilation):
```bash
# Para Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o whatsapp-viewer-linux .

# Para macOS (Apple Silicon M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o whatsapp-viewer-darwin-arm64 .

# Para Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o whatsapp-viewer.exe .
```

---

### Créditos e Agradecimentos

* **Andreas Mausch** ([@andreas-mausch](https://github.com/andreas-mausch)): Criador do projeto original [whatsapp-viewer](https://github.com/andreas-mausch/whatsapp-viewer) em C++.
* **TripCode & EliteAndroidApps**: Engenharia reversa e implementação do formato crypt12.
* **@torsade & colaboradores**: Contribuições para o suporte ao crypt14.
* **modernc.org/sqlite**: Driver SQLite em Go puro que viabiliza a portabilidade sem CGO.

---

### Licença

Distribuído sob a licença [MIT](LICENSE).
