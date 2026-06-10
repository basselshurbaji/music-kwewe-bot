package bot

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

var (
	adjectives = []string{
		"angry", "sleepy", "hungry", "sneaky", "grumpy", "spooky", "lazy", "chunky",
		"fluffy", "fuzzy", "tiny", "giant", "silly", "clumsy", "cheeky", "quiet",
		"loud", "rusty", "dusty", "shiny", "soggy", "crispy", "spicy", "cold",
		"warm", "old", "young", "lost", "broken", "cursed", "haunted", "shy",
		"brave", "smug", "blind", "deaf", "bored", "tired", "nervous", "confused",
		"proud", "strange", "wild", "calm", "dark", "bright", "fast", "slow",
		"stinky", "greedy",
	}

	creatures = []string{
		"kwewe", "goblin", "wizard", "bard", "knight", "elf", "dwarf", "orc",
		"troll", "ghost", "dragon", "cat", "frog", "duck", "penguin", "monkey",
		"raccoon", "sloth", "panda", "robot", "ninja", "pirate", "cowboy", "chef",
		"intern", "nerd", "witch", "gnome", "giant", "fairy", "demon", "vampire",
		"zombie", "king", "jester", "monk", "thief", "priest", "mage", "rogue",
		"druid", "ranger", "golem", "imp", "sprite", "gremlin", "kobold", "paladin",
		"warlock", "lich",
	}

	actions = []string{
		"yodels", "naps", "sings", "raps", "dances", "trips", "snores", "hiccups",
		"sneezes", "burps", "honks", "toots", "bonks", "slaps", "drops", "spins",
		"falls", "runs", "hops", "skips", "jumps", "crawls", "slides", "rolls",
		"floats", "swims", "flies", "stumbles", "mumbles", "grumbles", "wobbles",
		"wiggles", "jiggles", "giggles", "cackles", "shrieks", "howls", "barks",
		"meows", "quacks", "croaks", "roars", "hisses", "squeaks", "chomps",
		"chugs", "sniffs", "wails", "belches", "rambles",
	}
)

// Generate returns a random three-word passphrase of the form
// "<adjective> <creature> <action>", e.g. "cursed kwewe yodels".
func Generate() string {
	return fmt.Sprintf("%s %s %s", pick(adjectives), pick(creatures), pick(actions))
}

func pick(words []string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
	if err != nil {
		panic(err)
	}
	return words[n.Int64()]
}
