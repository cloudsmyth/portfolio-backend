# Card Search Go

A fast, lightweight command-line tool for searching Magic: The Gathering cards using the Scryfall API.

## Overview

Card Search Go brings the power of Scryfall's comprehensive MTG database directly to your terminal. No more context switching to your browser—search for cards, check oracle text, view mana costs, and more, all from the command line.

Perfect for deck builders, developers, and anyone who lives in the terminal.

## Features

- **Fast searches** - Query 25,000+ MTG cards instantly
- **Terminal-native** - No browser required
- **Accurate results** - Powered by the Scryfall API
- **Lightweight** - Minimal dependencies, fast execution
- **Cross-platform** - Works on Windows, macOS, and Linux

## Installation

### From Source

```bash
git clone https://github.com/cloudsmyth/card-search-go.git
cd card-search-go
go build
```

### Prerequisites

- Go 1.16 or higher
- Internet connection (to access Scryfall API)

## Usage

```bash
./card-search-go
```
This will start the program and drop you into the BubbleTea TUI experience.

## Future Improvements

We're planning several exciting enhancements:

- **Local caching** - Store frequently searched cards for offline access
- **Advanced filters** - Search by color, mana cost, card type, and more
- **Price integration** - Display current market prices
- **Format legality** - Quick format legality checks
- **Pokémon card support** - Expanding beyond MTG to support Pokémon TCG searches
- **Image display** - View card images in supported terminals

## Contributing

Contributions are welcome! Feel free to:

- Report bugs
- Suggest new features
- Submit pull requests
- Improve documentation

## License

This project is open source and available under the MIT License.

## Acknowledgments

- Built with [Go](https://golang.org/)
- Powered by [Scryfall API](https://scryfall.com/docs/api)
- Thanks to the MTG community

## Links

- [Scryfall](https://scryfall.com/) - The best MTG card database
- [Magic: The Gathering](https://magic.wizards.com/)

---

**Note**: This tool is not affiliated with or endorsed by Wizards of the Coast or Scryfall.
