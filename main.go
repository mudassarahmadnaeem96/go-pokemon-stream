package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-resty/resty/v2"
)

const (
	pokeAPIBase  = "https://pokeapi.co/api/v2/pokemon"
	maxPokemonID = 1010
)

type Pokemon struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Height  int        `json:"height"`
	Weight  int        `json:"weight"`
	Types   []TypeSlot `json:"types"`
	Stats   []StatSlot `json:"stats"`
	Sprites Sprites    `json:"sprites"`
}

type TypeSlot struct {
	Type TypeInfo `json:"type"`
}

type TypeInfo struct {
	Name string `json:"name"`
}

type StatSlot struct {
	BaseStat int      `json:"base_stat"`
	Stat     StatInfo `json:"stat"`
}

type StatInfo struct {
	Name string `json:"name"`
}

type Sprites struct {
	FrontDefault string       `json:"front_default"`
	Other        OtherSprites `json:"other"`
}

type OtherSprites struct {
	OfficialArtwork ArtworkSprite `json:"official-artwork"`
}

type ArtworkSprite struct {
	FrontDefault string `json:"front_default"`
}

type PokemonEvent struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Types      []string `json:"types"`
	Height     float64  `json:"height"`
	Weight     float64  `json:"weight"`
	HP         int      `json:"hp"`
	Attack     int      `json:"attack"`
	Defense    int      `json:"defense"`
	Speed      int      `json:"speed"`
	SpAttack   int      `json:"sp_attack"`
	SpDefense  int      `json:"sp_defense"`
	ImageURL   string   `json:"image_url"`
	SpriteURL  string   `json:"sprite_url"`
	TotalStats int      `json:"total_stats"`
}

var client *resty.Client

func init() {
	rand.Seed(time.Now().UnixNano())
	client = resty.New().SetTimeout(15 * time.Second)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

func fetchRandomPokemon() (*PokemonEvent, error) {
	id := rand.Intn(maxPokemonID) + 1
	url := fmt.Sprintf("%s/%d", pokeAPIBase, id)

	var pokemon Pokemon
	resp, err := client.R().SetResult(&pokemon).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode())
	}

	types := make([]string, len(pokemon.Types))
	for i, t := range pokemon.Types {
		types[i] = capitalize(t.Type.Name)
	}

	event := &PokemonEvent{
		ID:        pokemon.ID,
		Name:      capitalize(pokemon.Name),
		Types:     types,
		Height:    float64(pokemon.Height) / 10.0,
		Weight:    float64(pokemon.Weight) / 10.0,
		ImageURL:  pokemon.Sprites.Other.OfficialArtwork.FrontDefault,
		SpriteURL: pokemon.Sprites.FrontDefault,
	}

	if event.ImageURL == "" {
		event.ImageURL = pokemon.Sprites.FrontDefault
	}

	for _, stat := range pokemon.Stats {
		switch stat.Stat.Name {
		case "hp":
			event.HP = stat.BaseStat
		case "attack":
			event.Attack = stat.BaseStat
		case "defense":
			event.Defense = stat.BaseStat
		case "speed":
			event.Speed = stat.BaseStat
		case "special-attack":
			event.SpAttack = stat.BaseStat
		case "special-defense":
			event.SpDefense = stat.BaseStat
		}
	}

	event.TotalStats = event.HP + event.Attack + event.Defense + event.Speed + event.SpAttack + event.SpDefense

	return event, nil
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(6 * time.Second)
	defer ticker.Stop()

	sendPokemon := func() {
		pokemon, err := fetchRandomPokemon()
		if err != nil {
			log.Printf("Error fetching pokemon: %v", err)
			fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Failed to fetch Pokemon\"}\n\n")
			flusher.Flush()
			return
		}
		data, _ := json.Marshal(pokemon)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	sendPokemon()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			sendPokemon()
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"pokemon-stream"}`))
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Get("/", indexHandler)
	r.Get("/events", sseHandler)
	r.Get("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("⚡ Pokemon Stream starting on port %s", port)
	log.Printf("🎮 Open http://localhost:%s in your browser", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pokemon Stream</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
    <style>
        .font-pixel { font-family: 'Press Start 2P', cursive; }
        .pokeball-spin { animation: spin 1s linear infinite; }
        @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
        @keyframes float { 0%, 100% { transform: translateY(0px); } 50% { transform: translateY(-10px); } }
        .float { animation: float 3s ease-in-out infinite; }
        @keyframes fadeIn { from { opacity: 0; transform: scale(0.9); } to { opacity: 1; transform: scale(1); } }
        .fade-in { animation: fadeIn 0.5s ease-out; }
        .type-normal { background: linear-gradient(135deg, #A8A878, #8A8A59); }
        .type-fire { background: linear-gradient(135deg, #F08030, #DD6610); }
        .type-water { background: linear-gradient(135deg, #6890F0, #4A6FC4); }
        .type-electric { background: linear-gradient(135deg, #F8D030, #E0B000); }
        .type-grass { background: linear-gradient(135deg, #78C850, #5CA935); }
        .type-ice { background: linear-gradient(135deg, #98D8D8, #69C6C6); }
        .type-fighting { background: linear-gradient(135deg, #C03028, #9D2721); }
        .type-poison { background: linear-gradient(135deg, #A040A0, #803380); }
        .type-ground { background: linear-gradient(135deg, #E0C068, #C4A24B); }
        .type-flying { background: linear-gradient(135deg, #A890F0, #8A6EDB); }
        .type-psychic { background: linear-gradient(135deg, #F85888, #E33A65); }
        .type-bug { background: linear-gradient(135deg, #A8B820, #8B9A1B); }
        .type-rock { background: linear-gradient(135deg, #B8A038, #93802D); }
        .type-ghost { background: linear-gradient(135deg, #705898, #554374); }
        .type-dragon { background: linear-gradient(135deg, #7038F8, #4C08EF); }
        .type-dark { background: linear-gradient(135deg, #705848, #513F34); }
        .type-steel { background: linear-gradient(135deg, #B8B8D0, #9797BA); }
        .type-fairy { background: linear-gradient(135deg, #EE99AC, #E76991); }
        .stat-bar { transition: width 0.8s ease-out; }
        .card-glow { box-shadow: 0 0 30px rgba(251, 191, 36, 0.3); }
    </style>
</head>
<body class="bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <header class="text-center mb-8">
            <h1 class="font-pixel text-2xl md:text-4xl text-yellow-400 mb-4 drop-shadow-lg">
                ⚡ POKEMON STREAM
            </h1>
            <p class="text-purple-300 text-sm md:text-base">Real-time Pokemon discovery powered by PokeAPI</p>
            <div id="status" class="mt-4 flex items-center justify-center gap-2">
                <span class="relative flex h-3 w-3">
                    <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                    <span class="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                </span>
                <span class="text-green-400 text-sm">Live Stream Active</span>
            </div>
        </header>

        <div id="loading" class="flex flex-col items-center justify-center py-20">
            <div class="w-16 h-16 border-4 border-yellow-400 border-t-transparent rounded-full pokeball-spin"></div>
            <p class="text-yellow-400 mt-4 font-pixel text-xs">Catching Pokemon...</p>
        </div>

        <div id="pokemon-card" class="hidden max-w-2xl mx-auto">
            <div class="bg-gradient-to-br from-slate-800/90 to-slate-900/90 rounded-3xl p-6 md:p-8 backdrop-blur-sm border border-yellow-400/30 card-glow fade-in">
                <div class="flex flex-col md:flex-row gap-6 items-center">
                    <div class="relative">
                        <div class="absolute inset-0 bg-gradient-to-br from-yellow-400/20 to-purple-600/20 rounded-full blur-2xl"></div>
                        <img id="pokemon-image" src="" alt="Pokemon" class="relative w-48 h-48 md:w-56 md:h-56 object-contain float drop-shadow-2xl">
                    </div>
                    
                    <div class="flex-1 text-center md:text-left">
                        <div class="flex items-center justify-center md:justify-start gap-3 mb-2">
                            <span id="pokemon-id" class="text-slate-400 text-sm font-mono">#000</span>
                            <h2 id="pokemon-name" class="font-pixel text-xl md:text-2xl text-white"></h2>
                        </div>
                        
                        <div id="pokemon-types" class="flex gap-2 justify-center md:justify-start mb-4"></div>
                        
                        <div class="grid grid-cols-2 gap-4 text-sm mb-4">
                            <div class="bg-slate-700/50 rounded-lg p-3">
                                <span class="text-slate-400 block">Height</span>
                                <span id="pokemon-height" class="text-white font-bold">0 m</span>
                            </div>
                            <div class="bg-slate-700/50 rounded-lg p-3">
                                <span class="text-slate-400 block">Weight</span>
                                <span id="pokemon-weight" class="text-white font-bold">0 kg</span>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="mt-6 space-y-3">
                    <h3 class="text-yellow-400 font-pixel text-xs mb-4">BASE STATS</h3>
                    
                    <div class="grid gap-2">
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">HP</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-hp" class="stat-bar h-full bg-gradient-to-r from-green-500 to-green-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-hp-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">Attack</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-attack" class="stat-bar h-full bg-gradient-to-r from-red-500 to-orange-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-attack-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">Defense</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-defense" class="stat-bar h-full bg-gradient-to-r from-blue-500 to-blue-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-defense-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">Sp. Attack</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-sp-attack" class="stat-bar h-full bg-gradient-to-r from-purple-500 to-pink-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-sp-attack-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">Sp. Defense</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-sp-defense" class="stat-bar h-full bg-gradient-to-r from-teal-500 to-cyan-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-sp-defense-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                        <div class="flex items-center gap-3">
                            <span class="w-24 text-xs text-slate-400">Speed</span>
                            <div class="flex-1 bg-slate-700 rounded-full h-4 overflow-hidden">
                                <div id="stat-speed" class="stat-bar h-full bg-gradient-to-r from-yellow-500 to-amber-400 rounded-full" style="width: 0%"></div>
                            </div>
                            <span id="stat-speed-val" class="w-10 text-right text-xs text-white font-mono">0</span>
                        </div>
                    </div>

                    <div class="mt-4 pt-4 border-t border-slate-700 flex justify-between items-center">
                        <span class="text-slate-400 text-sm">Total Stats</span>
                        <span id="stat-total" class="text-yellow-400 font-pixel text-sm">0</span>
                    </div>
                </div>
            </div>
        </div>

        <div id="history" class="mt-8 max-w-2xl mx-auto">
            <h3 class="text-purple-300 text-sm mb-4 text-center">Recent Discoveries</h3>
            <div id="history-list" class="flex flex-wrap gap-2 justify-center"></div>
        </div>

        <div id="counter" class="text-center mt-8 text-slate-500 text-sm">
            Pokemon discovered: <span id="count" class="text-yellow-400 font-mono">0</span>
        </div>
    </div>

    <footer class="text-center py-6 text-slate-500 text-xs">
        Powered by <a href="https://pokeapi.co" class="text-purple-400 hover:text-purple-300" target="_blank">PokeAPI</a>
        • New Pokemon every 6 seconds
    </footer>

    <script>
        const maxStat = 255;
        let count = 0;
        const history = [];
        const maxHistory = 12;

        function getTypeClass(type) {
            const types = {
                'Normal': 'type-normal', 'Fire': 'type-fire', 'Water': 'type-water',
                'Electric': 'type-electric', 'Grass': 'type-grass', 'Ice': 'type-ice',
                'Fighting': 'type-fighting', 'Poison': 'type-poison', 'Ground': 'type-ground',
                'Flying': 'type-flying', 'Psychic': 'type-psychic', 'Bug': 'type-bug',
                'Rock': 'type-rock', 'Ghost': 'type-ghost', 'Dragon': 'type-dragon',
                'Dark': 'type-dark', 'Steel': 'type-steel', 'Fairy': 'type-fairy'
            };
            return types[type] || 'bg-slate-500';
        }

        function updateStat(id, value) {
            const bar = document.getElementById('stat-' + id);
            const val = document.getElementById('stat-' + id + '-val');
            const percent = Math.min((value / maxStat) * 100, 100);
            bar.style.width = percent + '%';
            val.textContent = value;
        }

        function updatePokemon(data) {
            document.getElementById('loading').classList.add('hidden');
            const card = document.getElementById('pokemon-card');
            card.classList.remove('hidden');
            card.querySelector('.fade-in').classList.remove('fade-in');
            void card.offsetWidth;
            card.querySelector('div').classList.add('fade-in');

            document.getElementById('pokemon-id').textContent = '#' + String(data.id).padStart(3, '0');
            document.getElementById('pokemon-name').textContent = data.name;
            document.getElementById('pokemon-image').src = data.image_url || data.sprite_url || '';
            document.getElementById('pokemon-height').textContent = data.height.toFixed(1) + ' m';
            document.getElementById('pokemon-weight').textContent = data.weight.toFixed(1) + ' kg';

            const typesContainer = document.getElementById('pokemon-types');
            typesContainer.innerHTML = data.types.map(type => 
                '<span class="px-3 py-1 rounded-full text-white text-xs font-bold shadow-lg ' + getTypeClass(type) + '">' + type + '</span>'
            ).join('');

            updateStat('hp', data.hp);
            updateStat('attack', data.attack);
            updateStat('defense', data.defense);
            updateStat('sp-attack', data.sp_attack);
            updateStat('sp-defense', data.sp_defense);
            updateStat('speed', data.speed);
            document.getElementById('stat-total').textContent = data.total_stats;

            count++;
            document.getElementById('count').textContent = count;

            if (!history.find(p => p.id === data.id)) {
                history.unshift({ id: data.id, name: data.name, sprite: data.sprite_url, type: data.types[0] });
                if (history.length > maxHistory) history.pop();
                updateHistory();
            }
        }

        function updateHistory() {
            const list = document.getElementById('history-list');
            list.innerHTML = history.map(p => 
                '<div class="w-12 h-12 rounded-full ' + getTypeClass(p.type) + ' p-1 opacity-70 hover:opacity-100 transition-opacity" title="' + p.name + '">' +
                '<img src="' + (p.sprite || '') + '" class="w-full h-full object-contain" alt="' + p.name + '">' +
                '</div>'
            ).join('');
        }

        function connect() {
            const evtSource = new EventSource('/events');
            
            evtSource.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    updatePokemon(data);
                } catch (e) {
                    console.error('Parse error:', e);
                }
            };

            evtSource.onerror = function() {
                document.getElementById('status').innerHTML = 
                    '<span class="relative flex h-3 w-3"><span class="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span></span>' +
                    '<span class="text-red-400 text-sm">Reconnecting...</span>';
                evtSource.close();
                setTimeout(connect, 3000);
            };
        }

        connect();
    </script>
</body>
</html>`
