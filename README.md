# ATS-TUI 🎯

> **A blazing-fast, keyboard-driven Applicant Tracking System (ATS) terminal user interface built with Go, Charm Bubble Tea, Lip Gloss, and PostgreSQL.**

[![Release](https://img.shields.io/github/v/release/NoahTek/ats-tui?color=00D2FF&label=Release)](https://github.com/NoahTek/ats-tui/releases)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14%2B-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Bubble Tea](https://img.shields.io/badge/Built%20With-Bubble%20Tea-00D2FF?style=flat)](https://github.com/charmbracelet/bubbletea)
[![Lip Gloss](https://img.shields.io/badge/Styled%20With-Lip%20Gloss-FF79C6?style=flat)](https://github.com/charmbracelet/lipgloss)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## ⚡ Overview

**ATS-TUI** brings modern recruiting workflows directly into your terminal. Designed for technical recruiters, engineering hiring managers, and headhunters who prefer staying inside the command line, ATS-TUI replaces bloated web interfaces with an ultra-responsive, distraction-free environment.

Built on Charm’s **Bubble Tea (Elm architecture)** and backed by **PostgreSQL** with high-performance connection pooling via `pgx/v5`, ATS-TUI provides full pipeline visibility, candidate search, interview stage management, and relational CRUD operations without ever touching a mouse.

---

## ✨ Features

- 🧙 **Zero-Manual-SQL Setup Wizard**:
  - Interactive first-run setup right inside the terminal.
  - Automatically creates the database (`CREATE DATABASE ats;`), runs embedded schema migrations, and seeds demo candidates & jobs.
  - Saves your configuration locally so subsequent launches connect instantly.
- 📑 **4-View Unified Workspace**:
  - `[1] Profiles`: Candidate repository with contact info, qualifications, and searchable resumes.
  - `[2] Applications`: Dual-mode pipeline tracking (interactive Kanban board & wide table view).
  - `[3] Jobs`: Job postings, requisitions, departments, client associations, and headcount metrics.
  - `[4] Organisations`: Client companies, industry classifications, and sub-sectors.
- 📋 **Interactive 4-Stage Kanban Pipeline**:
  - Visual stage columns: **Interested ➔ Follow-Up ➔ Call-Back ➔ Closed**.
  - One-key candidate stage transitions (`m` to advance, `b` to regress) synchronized directly to PostgreSQL.
- 🔍 **Full-Text & In-Memory Multi-Column Filtering**:
  - PostgreSQL `tsvector` + GIN-indexed full-text search across candidate profiles and resume contents.
  - Interactive column filter modal (`f`) with value checkboxes and multi-criteria drill-downs.
- 👁️ **Dual-Pane Inspector & Modal Viewport**:
  - Live split pane (`p`) showing comprehensive candidate summaries and interview notes in real time.
  - Dedicated full-screen inspection modal (`Enter` / `d`) with smooth vertical scrolling.
- 🛠️ **Full Terminal CRUD**:
  - In-terminal popup modal forms for creating (`a`) and editing (`e`) Candidates, Applications, Jobs, and Clients.
  - Safe deletion workflow (`x` / `D`) with confirmation prompts and cascade protection.
- ↔️ **Universal Horizontal Column Slicing**:
  - Never suffer from terminal text wrapping or misaligned ASCII columns. Scroll table columns horizontally (`[` / `]`) seamlessly regardless of terminal width.

---

## 🚀 Getting Started

### Prerequisites

| Setup Path | Requirements |
| :--- | :--- |
| **Running Pre-compiled Binary** | **PostgreSQL 14+** (local, Docker, or remote). *No Go installation required.* |
| **Building from Source** | **Go 1.22+** and **PostgreSQL 14+** |

---

### Option 1: Quick Install via Pre-compiled Binary (Recommended)

Download the latest release binary for Linux, macOS, or Windows from the [GitHub Releases](https://github.com/NoahTek/ats-tui/releases) (or Gitea Releases).

#### Linux & macOS
```bash
# 1. Download the release binary (or grab from the Releases page)
curl -LO https://github.com/NoahTek/ats-tui/releases/latest/download/ats-tui

# 2. Make it executable
chmod +x ats-tui

# 3. Launch (the Setup Wizard handles the rest!)
./ats-tui
```

#### Windows
1. Download `ats-tui-windows-amd64.exe` (or rename the binary to `ats-tui.exe`).
2. Run in **PowerShell** or **Command Prompt**:
```powershell
.\ats-tui.exe
```
> **Tip for Windows SmartScreen**: If Windows displays *"Windows protected your PC"* on first launch, click **More info** ➔ **Run anyway** (or right-click `ats-tui.exe` ➔ **Properties** ➔ check **Unblock** ➔ click **Apply**).

---

### Option 2: Build from Source (For Developers)

```bash
# 1. Clone repository
git clone https://github.com/NoahTek/ats-tui.git
cd ats-tui

# 2. Build the binary
go build -o ats-tui .

# 3. Launch
./ats-tui
```

---

## 🧙 In-App Database Setup Wizard

When you launch `ats-tui` for the first time without an active connection, the **In-App Setup Wizard** pops up automatically:

```text
 ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
 ┃               🎯  ATS-TUI Database Setup Wizard             ┃
 ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
 Welcome! Configure your PostgreSQL connection below.
 Settings will be saved to your local config automatically.

 Host:            localhost
 Port:            5432
 Database Name:   ats
 Username:        postgres
 Password:        ••••••••
 SSL Mode:        disable
 Seed Demo:       [✓] Load demo candidates & pipeline data

 [ Connect & Launch ATS ]
```

### What happens when you connect:
1. **Auto Database Creation**: The app checks if database `ats` exists; if not, it connects to PostgreSQL and creates it automatically.
2. **Auto Schema Migrations**: All tables, enums, sequences, and full-text indexes are automatically created.
3. **Auto Seed**: If selected, realistic demo candidates, jobs, client companies, and applications are populated.
4. **Auto Save**: Settings are saved locally so you never have to re-enter them.

> **Tip:** You can re-open the setup wizard at any time by running:
> ```bash
> ./ats-tui -setup
> ```

---

## 🐘 PostgreSQL Quickstart Tips

If you are setting up a fresh PostgreSQL installation on Linux or inside Docker:

### 1. Fresh Debian / Ubuntu Installations
Debian/Ubuntu installs PostgreSQL with no password assigned to user `postgres`, which causes password authentication to fail over TCP. Set a password with this one-liner:
```bash
su - postgres -c "psql -c \"ALTER USER postgres WITH PASSWORD 'postgres';\""
```

### 2. Using Docker (Instant Setup)
You can run a local PostgreSQL instance with a single command:
```bash
docker run -d --name ats-postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
```

### 3. Password-Free (Trust Mode)
If you prefer not using a password, set `127.0.0.1/32` to `trust` in `/etc/postgresql/*/main/pg_hba.conf` and leave the Password field blank in the ATS-TUI setup wizard.

---

## ⚙️ Configuration File Locations

ATS-TUI stores its persistent configuration file in the standard OS configuration directory:

| Operating System | Location |
| :--- | :--- |
| **Linux** | `~/.config/ats-tui/config.json` |
| **macOS** | `~/Library/Application Support/ats-tui/config.json` |
| **Windows** | `%APPDATA%\ats-tui\config.json`<br>`C:\Users\<Username>\AppData\Roaming\ats-tui\config.json` |

To reset your configuration back to clean install state, delete this file or pass `./ats-tui -setup`.

### CLI / Environment Variable Overrides
You can also override settings via CLI flags or environment variables:
```bash
# CLI flag
./ats-tui -db "postgres://user:password@localhost:5432/ats"

# Environment Variable
export DATABASE_URL="postgres://user:password@localhost:5432/ats"
./ats-tui
```

---

## ⌨️ Keybindings & Navigation

### Global & Tabs
| Key | Action |
| :--- | :--- |
| `1` / `2` / `3` / `4` | Switch directly to **Profiles**, **Applications**, **Jobs**, or **Organisations** |
| `Tab` / `Shift+Tab` | Cycle forward / backward through workspace tabs |
| `q` / `Ctrl+C` | Quit ATS-TUI |
| `R` / `Ctrl+R` | Hard reload all datasets from PostgreSQL |

### Applications Pipeline (Kanban & Table)
| Key | Action |
| :--- | :--- |
| `v` or `t` | Toggle between **4-Stage Kanban Board** and **Full Table View** |
| `h` / `l` or `←` / `→` | Navigate left / right between Kanban stage columns |
| `j` / `k` or `↑` / `↓` | Move up / down through application cards in the active column |
| `m` | **Move candidate forward** to next stage (*Interested ➔ Follow-Up ➔ Call-Back ➔ Closed*) |
| `b` | **Move candidate backward** to previous stage |

### Search, Sort & Filters
| Key | Action |
| :--- | :--- |
| `/` | Open global search bar (`Enter` to submit, `Esc` to cancel) |
| `f` | Open **Column Filter Modal** (toggle criteria with `Space`, search values with `/`) |
| `r` | Reset all search queries and active column filters |
| `s` | Cycle active sort column |
| `S` (Shift+S) | Toggle sort direction (**ASC ⇄ DESC**) |
| `[` / `]` | Scroll visible table columns left / right |

### Inspection & CRUD Actions
| Key | Action |
| :--- | :--- |
| `p` | Toggle split-pane inspection drawer at the bottom of the screen |
| `Enter` / `d` | Open full detailed inspection modal with interview notes history |
| `a` | **Add** a new entity (Candidate, Application, Job, or Client) |
| `e` | **Edit** the currently highlighted entity |
| `x` / `D` / `Del` | **Delete** highlighted entity (requires `y`/`n` confirmation) |
| `Esc` | Close open modal, cancel search, or dismiss forms |

---

## 🗄️ Database Architecture

ATS-TUI uses a clean relational schema in PostgreSQL:

```
  ┌───────────────┐          ┌───────────────┐          ┌──────────────┐
  │  candidates   │◄─────────┤ applications  ├─────────►│     jobs     │
  │ (CAN-001...)  │          │(stage status) │          │  (JOB-001)   │
  └───────┬───────┘          └───────┬───────┘          └──────┬───────┘
          │                          │                         │
          ▼                          ▼                         ▼
   ┌─────────────┐             ┌───────────┐             ┌───────────┐
   │  locations  │             │   notes   │             │  clients  │
   │educational_q│             │(interview)│             │(org_code) │
   └─────────────┘             └─────┬─────┘             └─────┬─────┘
                                     ▼                         ▼
                              ┌─────────────┐            ┌───────────┐
                              │  products   │            │ industry  │
                              └─────────────┘            └───────────┘
```

---

## 🧪 Testing

Run the full automated test suite across database migrations, UI components, Setup Wizard, and data models:

```bash
go test -v ./...
```

---

## 📂 Project Structure

```text
ats-tui/
├── db/
│   ├── config.go              # Local configuration loader & saver (~/.config/ats-tui)
│   ├── config_test.go         # Tests for config serialization & resolution
│   ├── db.go                  # pgx connection pool and migration runner
│   ├── seed.go                # Embedded demo data seeder & empty DB detector
│   ├── seed.sql               # Comprehensive initial demo data
│   └── migrations/            # Auto-applied SQL migrations (000, 001, 002)
├── kanban/
│   ├── kanban.go              # Bubble Tea 4-stage pipeline component
│   └── kanban_test.go         # Pipeline state transition tests
├── models/
│   ├── application.go         # Application & Kanban card structures
│   ├── candidate.go           # Candidate profile queries & tsvector search
│   ├── job.go                 # Job posting models and operations
│   ├── organisation.go        # Client companies & industry models
│   ├── note.go                # Interview feedback & application notes
│   └── lookup.go              # Lookup tables (locations, skills, degrees)
├── ui/
│   ├── root.go                # Root coordinator model (Setup <-> App)
│   ├── setup.go               # In-App Database Setup Wizard UI
│   ├── setup_test.go          # Tests for Setup wizard form navigation
│   ├── app.go                 # Main Bubble Tea workspace coordinator
│   ├── form.go                # Dynamic CRUD popup modal forms
│   └── app_test.go            # UI navigation and lifecycle tests
├── go.mod                     # Go module dependencies
├── main.go                    # Entrypoint & startup resolution
└── README.md                  # Documentation
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to open an issue or submit a pull request on [GitHub](https://github.com/NoahTek/ats-tui) or your local Gitea instance.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
