# ⚡ Pokemon Stream

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![SSE](https://img.shields.io/badge/SSE-Real--time-FF6B6B?style=for-the-badge&logo=lightning&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind-CSS-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)
![PokeAPI](https://img.shields.io/badge/PokeAPI-Powered-FFCB05?style=for-the-badge&logo=pokemon&logoColor=black)
![Render](https://img.shields.io/badge/Render-Deployed-46E3B7?style=for-the-badge&logo=render&logoColor=white)

Real-time Pokemon discovery stream using Server-Sent Events. Watch random Pokemon appear with their stats, types, and official artwork!

## ✨ Features

- 🔴 **Live SSE Stream** — New Pokemon every 6 seconds
- 📊 **Full Stats Display** — HP, Attack, Defense, Speed & more
- 🎨 **Type Badges** — Color-coded Pokemon types
- 📜 **Discovery History** — Track recently discovered Pokemon
- 📱 **Responsive Design** — Works on all devices
- 🎮 **Retro Pixel Font** — Authentic Pokemon aesthetic
- ⚡ **Zero Config** — Just run and enjoy

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go + Chi Router |
| HTTP Client | Resty |
| Frontend | Tailwind CSS |
| Streaming | Server-Sent Events |
| Data | PokeAPI |

## 🚀 Quick Start

Clone the repository:

```bash
git clone https://github.com/smart-developer1791/go-pokemon-stream
cd go-pokemon-stream
```

Initialize dependencies and run:

```bash
go mod tidy
go run .
```

Open browser:

```text
http://localhost:8080
```

## 📡 API Endpoints

| Endpoint | Description |
|----------|-------------|
| GET / | Main web interface |
| GET /events | SSE stream of Pokemon |
| GET /health | Health check endpoint |

## 📦 Dependencies

```text
github.com/go-chi/chi/v5     — Lightweight router
github.com/go-resty/resty/v2 — HTTP client
```

## 🎯 SSE Event Format

```json
{
  "id": 25,
  "name": "Pikachu",
  "types": ["Electric"],
  "height": 0.4,
  "weight": 6.0,
  "hp": 35,
  "attack": 55,
  "defense": 40,
  "speed": 90,
  "sp_attack": 50,
  "sp_defense": 50,
  "total_stats": 320,
  "image_url": "https://..."
}
```

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |

## 📊 Project Structure

```text
go-pokemon-stream/
├── main.go         # Server, SSE handler, HTML
├── go.mod          # Go module definition
├── render.yaml     # Render deployment config
├── .gitignore      # Git ignore rules
└── README.md       # Documentation
```

## 🌐 PokeAPI

This project uses the free [PokeAPI](https://pokeapi.co/) which provides:

- 1000+ Pokemon
- Official artwork & sprites
- Complete stats data
- Type information

No API key required!

## 🎮 Pokemon Stats Guide

| Stat | Max Value | Description |
|------|-----------|-------------|
| HP | 255 | Hit Points |
| Attack | 255 | Physical attack |
| Defense | 255 | Physical defense |
| Sp. Attack | 255 | Special attack |
| Sp. Defense | 255 | Special defense |
| Speed | 255 | Move order priority |

---

## Deploy in 10 seconds

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)
