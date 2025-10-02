# Threadbound Desktop App

A native macOS application for generating beautiful books from iMessages.

## Architecture

- **Frontend**: React + TypeScript + Vite
- **Framework**: Tauri (Rust)
- **Backend**: Go API server (runs as sidecar process)
- **Communication**: Frontend → localhost:8765 → Go API

## Development

### Prerequisites

- Node.js (v16+)
- Rust & Cargo
- Go (v1.22+)

### Running in Dev Mode

```bash
# From the project root
./scripts/dev.sh
```

This will:
1. Build the Go backend binaries
2. Start the Tauri dev server
3. Launch the app with hot reload enabled

### Manual Dev Mode

```bash
# 1. Build the Go backend first
./scripts/build-backend.sh

# 2. Start Tauri dev mode
cd desktop
npm run tauri dev
```

## Building for Production

### Quick Build

```bash
# From project root
./scripts/build-app.sh
```

### Manual Build

```bash
# 1. Build Go backend
./scripts/build-backend.sh

# 2. Build Tauri app
cd desktop
npm run tauri build
```

The built app will be in: `desktop/src-tauri/target/release/bundle/`

## Project Structure

```
desktop/
├── src/                    # React frontend
│   ├── App.tsx            # Main UI component
│   ├── App.css            # Styles
│   └── main.tsx           # Entry point
├── src-tauri/             # Tauri Rust wrapper
│   ├── src/
│   │   ├── main.rs        # Rust entry point
│   │   └── lib.rs         # Tauri app config & sidecar setup
│   ├── binaries/          # Go backend binaries (built by scripts)
│   ├── tauri.conf.json    # Tauri configuration
│   └── Cargo.toml         # Rust dependencies
├── package.json
└── README.md
```

## How It Works

1. **App Launch**: Tauri starts and spawns the Go backend as a sidecar process
2. **Backend API**: Go server starts on `localhost:8765`
3. **Frontend**: React app connects to the API
4. **User Flow**:
   - User enters iMessage database path
   - Frontend sends request to `/api/generate`
   - Backend processes messages asynchronously
   - Frontend polls `/api/jobs/{id}` for status
   - When complete, user gets the generated file

## API Endpoints

- `GET /api/health` - Check backend status
- `POST /api/generate` - Start book generation job
- `GET /api/jobs/{id}` - Get job status
- `GET /api/jobs` - List all jobs

## Configuration

The app uses the same configuration as the CLI tool. See the main [README](../README.md) for details on:
- Database paths
- Attachments
- Output formats
- Contact names

## Distribution

The built `.app` bundle includes:
- React frontend
- Tauri runtime
- Go backend binary (auto-selected for architecture)

No external dependencies needed - everything is bundled!

## Troubleshooting

### Backend not connecting

1. Check that the Go binary exists in `desktop/src-tauri/binaries/`
2. Run `./scripts/build-backend.sh` to rebuild
3. Check the Tauri console for sidecar spawn errors

### Build fails

1. Ensure all prerequisites are installed
2. Try `cd desktop && npm install` to refresh dependencies
3. Clean build: `rm -rf desktop/src-tauri/target && ./scripts/build-app.sh`

## Next Steps

- [ ] Add file picker for database selection
- [ ] Add license key validation
- [ ] Add update mechanism
- [ ] Create app icons
- [ ] Set up code signing for macOS distribution
