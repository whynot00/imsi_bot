package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/whynot00/imsi-bot/internal/config"

	"github.com/whynot00/imsi-bot/internal/parser"
)

func main() {
	// Load .env before anything else.
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	database, err := db.Connect(cfg.Postgres)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()
	fmt.Fprintln(os.Stderr, "connected to postgres")

	inPath := flag.String("in", "", "path to source CSV (required)")
	kind := flag.String("kind", "", "file type: parametr or rk (required)")
	outDir := flag.String("out", "output", "directory for output CSVs")
	encoding := flag.String("encoding", "utf8", "source encoding: utf8 (default) or cp1251")
	flag.Parse()

	if *inPath == "" || *kind == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	f, err := os.Open(*inPath)
	if err != nil {
		log.Fatalf("open input: %v", err)
	}
	defer f.Close()

	var src io.Reader = f
	if *encoding == "cp1251" {
		src, err = windows1251Reader(f)
		if err != nil {
			log.Fatalf("encoding setup: %v", err)
		}
	}

	base := strings.TrimSuffix(filepath.Base(*inPath), filepath.Ext(*inPath))

	switch *kind {
	case "parametr":
		result, err := parser.Parse(src)
		if err != nil {
			log.Fatalf("parse: %v", err)
		}

		writeFile(filepath.Join(*outDir, base+"_devices.csv"), func(w io.Writer) error {
			return parser.WriteDevices(w, result.Devices)
		})
		writeFile(filepath.Join(*outDir, base+"_locations.csv"), func(w io.Writer) error {
			return parser.WriteLocations(w, result.Locations)
		})

		fmt.Fprintf(os.Stderr, "devices:   %d\n", len(result.Devices))
		fmt.Fprintf(os.Stderr, "locations: %d\n", len(result.Locations))

	case "rk":
		devices, err := parser.ParseRaw(src)
		if err != nil {
			log.Fatalf("parse rk: %v", err)
		}

		writeFile(filepath.Join(*outDir, base+"_rk.csv"), func(w io.Writer) error {
			return parser.WriteRawDevices(w, devices)
		})

		fmt.Fprintf(os.Stderr, "rk devices: %d\n", len(devices))

	default:
		log.Fatalf("unknown -kind %q: must be parametr or rk", *kind)
	}

	_ = database // repo usage will be wired here
}

func writeFile(path string, fn func(io.Writer) error) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := fn(f); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
	fmt.Fprintf(os.Stderr, "written → %s\n", path)
}
