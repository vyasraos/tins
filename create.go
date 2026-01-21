package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/spf13/cobra"
)

// generateInstanceName generates a Docker-style two-word name (adjective-noun)
func generateInstanceName() (string, error) {
	adjectives := []string{
		"admiring", "adoring", "agitated", "amazing", "angry", "awesome", "beautiful",
		"blissful", "bold", "boring", "brave", "busy", "charming", "clever", "cool",
		"compassionate", "competent", "condescending", "confident", "cranky", "crazy",
		"dazzling", "determined", "distracted", "dreamy", "eager", "ecstatic", "elastic",
		"elated", "elegant", "eloquent", "epic", "exciting", "fervent", "festive",
		"flamboyant", "focused", "friendly", "frosty", "funny", "gallant", "gifted",
		"goofy", "gracious", "great", "happy", "hardcore", "heuristic", "hopeful",
		"hungry", "infallible", "inspiring", "intelligent", "interesting", "jolly",
		"jovial", "keen", "kind", "laughing", "loving", "lucid", "magical", "mystical",
		"modest", "musing", "naughty", "nervous", "nice", "nifty", "nostalgic",
		"objective", "optimistic", "peaceful", "pedantic", "pensive", "practical",
		"priceless", "quirky", "quizzical", "recursing", "relaxed", "reverent",
		"romantic", "sad", "serene", "sharp", "silly", "sleepy", "stoic", "strange",
		"stupefied", "suspicious", "sweet", "tender", "thirsty", "trusting", "unruffled",
		"upbeat", "vibrant", "vigilant", "vigorous", "wizardly", "wonderful", "xenodochial",
		"youthful", "zealous", "zen",
	}

	nouns := []string{
		"albattani", "allen", "almeida", "agnesi", "archimedes", "ardinghelli", "aryabhata",
		"austin", "babbage", "banach", "banzai", "bardeen", "bartik", "bassi", "beaver",
		"bell", "benz", "bhabha", "bhaskara", "black", "blackburn", "blackwell", "bohr",
		"booth", "borg", "bose", "bouman", "boyd", "brahmagupta", "brattain", "brown",
		"buck", "burnell", "cannon", "carson", "cartwright", "carver", "cerf", "chandrasekhar",
		"chaplygin", "chatelet", "chatterjee", "chebyshev", "clarke", "cohen", "colden",
		"cori", "cray", "curie", "curran", "darwin", "davinci", "dewdney", "dhawan",
		"diffie", "dijkstra", "dirac", "driscoll", "dubinsky", "easley", "edison", "einstein",
		"elbakyan", "elgamal", "elion", "ellis", "engelbart", "euclid", "euler", "faraday",
		"feistel", "fermat", "fermi", "feynman", "franklin", "gagarin", "galileo", "galois",
		"ganguly", "gates", "gauss", "germain", "goldberg", "goldstine", "goldwasser",
		"golick", "goodall", "gould", "greider", "grothendieck", "haibt", "hamilton",
		"haslett", "hawking", "heisenberg", "hellman", "hermann", "herschel", "hertz",
		"heyrovsky", "hodgkin", "hofstadter", "hoover", "hopper", "hugle", "hypatia",
		"ishizaka", "jackson", "jang", "jemison", "jennings", "jepsen", "johnson", "joliot",
		"jones", "kalam", "kapitsa", "kare", "keldysh", "keller", "kepler", "khayyam",
		"khorana", "kilby", "kirch", "knuth", "kowalevski", "lalande", "lamarr", "lamport",
		"leakey", "leavitt", "lederberg", "lehmann", "lewin", "lichterman", "liskov",
		"lovelace", "lumiere", "mahavira", "margulis", "matsumoto", "maxwell", "mayer",
		"mccarthy", "mcclintock", "mclaren", "mclean", "mcnulty", "meitner", "mendel",
		"mendeleev", "meninsky", "merkle", "mestorf", "minsky", "mirzakhani", "moore",
		"morse", "murdock", "moser", "napier", "nash", "neumann", "newton", "nightingale",
		"nobel", "noether", "northcutt", "noyce", "panini", "pare", "pascal", "pasteur",
		"payne", "perlman", "pike", "poincare", "poitras", "proskuriakova", "ptolemy",
		"raman", "ramanujan", "ride", "montalcini", "ritchie", "rhodes", "robinson",
		"roentgen", "rosalind", "rubin", "saha", "sammet", "sanderson", "satoshi",
		"shamir", "shannon", "shaw", "shirley", "shockley", "shtern", "sinoussi",
		"snyder", "solomon", "spence", "stonebraker", "sutherland", "swanson", "swartz",
		"swirles", "taussig", "tereshkova", "tesla", "tharp", "thompson", "torvalds",
		"tu", "turing", "varahamihira", "vaughan", "visvesvaraya", "volhard", "villani",
		"wah", "wiles", "williams", "williamson", "wilson", "wing", "wozniak", "wright",
		"wu", "yalow", "yonath", "zhukovsky",
	}

	// Select random adjective
	adjIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	if err != nil {
		return "", fmt.Errorf("failed to generate random adjective: %w", err)
	}
	adjective := adjectives[adjIdx.Int64()]

	// Select random noun
	nounIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	if err != nil {
		return "", fmt.Errorf("failed to generate random noun: %w", err)
	}
	noun := nouns[nounIdx.Int64()]

	return fmt.Sprintf("%s-%s", adjective, noun), nil
}

var createCmd = &cobra.Command{
	Use:   "create [instance-name]",
	Short: "Create a new temporary instance",
	Long:  "Create a new ephemeral OpenStack instance with automatic SSH key generation. If instance-name is not provided, a random Docker-style two-word name (adjective-noun) will be generated.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var instanceName string
		var err error

		if len(args) > 0 && args[0] != "" {
			instanceName = args[0]
		} else {
			// Generate random Docker-style name (adjective-noun)
			instanceName, err = generateInstanceName()
			if err != nil {
				return fmt.Errorf("failed to generate instance name: %w", err)
			}
			fmt.Printf("Generated instance name: %s\n", instanceName)
		}

		fullInstanceName := fmt.Sprintf("%s%s", InstanceNamePrefix, instanceName)

		// Load configuration
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		// Create OpenStack client
		ctx := context.Background()
		client, err := NewOpenStackClient(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to create OpenStack client: %w", err)
		}

		// Generate SSH key pair
		fmt.Printf("Generating SSH key pair for %s...\n", fullInstanceName)
		keyPair, err := GenerateSSHKey(instanceName)
		if err != nil {
			return fmt.Errorf("failed to generate SSH key: %w", err)
		}
		fmt.Printf("SSH key pair created: %s\n", keyPair.PrivateKeyPath)

		// Create instance
		fmt.Printf("Creating instance %s...\n", fullInstanceName)
		server, err := client.CreateInstance(ctx, fullInstanceName, keyPair.PublicKey)
		if err != nil {
			// Clean up SSH keys on failure
			if deleteErr := DeleteSSHKey(instanceName); deleteErr != nil {
				// Log but don't fail on cleanup error
				fmt.Printf("Warning: Failed to clean up SSH key: %v\n", deleteErr)
			}
			return fmt.Errorf("failed to create instance: %w", err)
		}

		fmt.Printf("Instance created successfully!\n")
		fmt.Printf("  ID: %s\n", server.ID)
		fmt.Printf("  Name: %s\n", server.Name)
		fmt.Printf("  Status: %s\n", server.Status)

		// Wait for instance to become active
		fmt.Printf("Waiting for instance to become active...\n")
		timeout := 5 * time.Minute
		if err := client.WaitForInstanceActive(ctx, server.ID, timeout); err != nil {
			fmt.Printf("Warning: Instance may not be ready yet: %v\n", err)
		} else {
			fmt.Printf("Instance is now ACTIVE\n")
		}

		// Get updated server info to show IP addresses
		var instanceIP string
		server, err = client.GetInstance(ctx, server.ID)
		if err == nil {
			if len(server.Addresses) > 0 {
				fmt.Printf("\nInstance IP addresses:\n")
				for networkName, addrList := range server.Addresses {
					// Type assert to []interface{} and then extract address info
					if addresses, ok := addrList.([]interface{}); ok {
						for _, addrInterface := range addresses {
							if addrMap, ok := addrInterface.(map[string]interface{}); ok {
								if addrType, ok := addrMap["OS-EXT-IPS:type"].(string); ok {
									if addrType == "fixed" || addrType == "floating" {
										if addr, ok := addrMap["addr"].(string); ok {
											fmt.Printf("  %s: %s\n", networkName, addr)
											// Use the first IP address found for SSH connection
											if instanceIP == "" {
												instanceIP = addr
											}
										}
									}
								} else if addr, ok := addrMap["addr"].(string); ok {
									// Fallback: show address if type is not available
									fmt.Printf("  %s: %s\n", networkName, addr)
									// Use the first IP address found for SSH connection
									if instanceIP == "" {
										instanceIP = addr
									}
								}
							}
						}
					}
				}
			}
		}

		fmt.Printf("\nSSH connection:\n")
		if instanceIP != "" {
			fmt.Printf("  ssh -i %s root@%s\n", keyPair.PrivateKeyPath, instanceIP)
		} else {
			fmt.Printf("  ssh -i %s root@<instance-ip>\n", keyPair.PrivateKeyPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
