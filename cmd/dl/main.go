package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dl/pkg/config"
	"dl/pkg/downloader"
	"dl/pkg/remote"
)

//go:embed ../../VERSION
var version string

type Args struct {
	url        string
	debug      bool
	verbose    bool
	keepVideo  bool
	xrated     bool
	videoDest  string
	configPath string
}

func main() {
	args := parseArgs()

	// Check if yt-dlp is installed
	if err := downloader.CheckYtDlp(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(args.configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Determine media type and destination
	mediaType := downloader.MP3
	format := "mp3"
	var destination string
	var purpose remote.Purpose

	if args.keepVideo || args.videoDest != "" || args.xrated {
		mediaType = downloader.Video
		format = "mp4"
		purpose = remote.PurposeVideos

		if args.xrated {
			destination, err = cfg.GetXDestPath()
			purpose = remote.PurposeXRated
		} else if args.videoDest != "" {
			destination = args.videoDest
		} else {
			destination, err = cfg.GetVDestPath()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// MP3 download
		destination, err = cfg.GetMP3sPath()
		purpose = remote.PurposeMP3s
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if args.verbose {
		fmt.Printf("Downloading from: %s\n", args.url)
		fmt.Printf("Media type: %s\n", format)
		fmt.Printf("Destination: %s\n", destination)
		if len(cfg.ActiveRemotes) > 0 {
			fmt.Printf("Active remotes: %s\n", strings.Join(cfg.GetActiveRemoteNames(), ", "))
		}
		fmt.Println()
	}

	// Download media
	opts := &downloader.Options{
		URL:         args.url,
		Destination: destination,
		MediaType:   mediaType,
		Format:      format,
		Verbose:     args.verbose,
		XRated:      args.xrated,
	}

	dl := downloader.New(opts)
	localPath, err := dl.Download()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded to: %s\n", localPath)

	// Copy to remotes if configured
	if len(cfg.ActiveRemotes) > 0 {
		copier := remote.New(cfg, args.verbose)
		if err := copier.CopyToRemotes(localPath, purpose); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: some remote copies failed: %v\n", err)
		}
	}

	fmt.Println("Done!")
}

func parseArgs() *Args {
	args := &Args{}

	flag.StringVar(&args.configPath, "c", "~/dl.config", "Path to configuration file")
	flag.BoolVar(&args.debug, "d", false, "Enable debug mode (alias for verbose)")
	flag.BoolVar(&args.verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&args.keepVideo, "k", false, "Download and keep video")
	flag.BoolVar(&args.xrated, "x", false, "Download x-rated video to xdest")
	flag.StringVar(&args.videoDest, "V", "", "Download video to specified directory")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dl - Download videos and audio from various websites\n\n")
		fmt.Fprintf(os.Stderr, "Version: %s\n\n", strings.TrimSpace(version))
		fmt.Fprintf(os.Stderr, "Usage: %s [options] URL\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Downloads media from URLs using yt-dlp.\n")
		fmt.Fprintf(os.Stderr, "By default, downloads audio as MP3, unless -k, -x, or -V options are provided.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  Edit ~/dl.config to configure local and remote destinations.\n")
		fmt.Fprintf(os.Stderr, "  See README.md for configuration details.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  dl https://www.youtube.com/watch?v=dQw4w9WgXcQ\n")
		fmt.Fprintf(os.Stderr, "  dl -v https://www.youtube.com/watch?v=dQw4w9WgXcQ\n")
		fmt.Fprintf(os.Stderr, "  dl -k https://www.youtube.com/watch?v=dQw4w9WgXcQ\n")
		fmt.Fprintf(os.Stderr, "  dl -V ~/Videos https://www.youtube.com/watch?v=dQw4w9WgXcQ\n")
	}

	flag.Parse()

	// Get URL from remaining arguments
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: URL required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	args.url = flag.Arg(0)

	// Debug mode enables verbose
	if args.debug {
		args.verbose = true
	}

	return args
}
