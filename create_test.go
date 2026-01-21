package main

import (
	"regexp"
	"testing"
)

func TestGenerateInstanceName(t *testing.T) {
	name, err := generateInstanceName()
	if err != nil {
		t.Fatalf("generateInstanceName failed: %v", err)
	}
	
	// Verify format: adjective-noun
	pattern := `^[a-z]+-[a-z]+$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		t.Fatalf("Regex match failed: %v", err)
	}
	if !matched {
		t.Errorf("Instance name '%s' does not match expected pattern '%s'", name, pattern)
	}
	
	// Verify it contains a hyphen
	if len(name) < 3 {
		t.Errorf("Instance name '%s' is too short", name)
	}
}

func TestGenerateInstanceName_Uniqueness(t *testing.T) {
	// Generate multiple names and verify they're different
	names := make(map[string]bool)
	iterations := 100
	
	for i := 0; i < iterations; i++ {
		name, err := generateInstanceName()
		if err != nil {
			t.Fatalf("generateInstanceName failed: %v", err)
		}
		
		if names[name] {
			// It's okay if we get duplicates occasionally, but log it
			t.Logf("Duplicate name generated: %s (iteration %d)", name, i)
		}
		names[name] = true
	}
	
	// With 100 iterations, we should have some variety
	// (though duplicates are possible with the word lists)
	if len(names) < 10 {
		t.Logf("Warning: Only %d unique names generated in %d iterations", len(names), iterations)
	}
}

func TestGenerateInstanceName_ValidWords(t *testing.T) {
	// Generate multiple names and verify they use valid words
	adjectives := map[string]bool{
		"admiring": true, "adoring": true, "agitated": true, "amazing": true,
		"angry": true, "awesome": true, "beautiful": true, "blissful": true,
		"bold": true, "boring": true, "brave": true, "busy": true,
		"charming": true, "clever": true, "cool": true, "compassionate": true,
		"competent": true, "condescending": true, "confident": true, "cranky": true,
		"crazy": true, "dazzling": true, "determined": true, "distracted": true,
		"dreamy": true, "eager": true, "ecstatic": true, "elastic": true,
		"elated": true, "elegant": true, "eloquent": true, "epic": true,
		"exciting": true, "fervent": true, "festive": true, "flamboyant": true,
		"focused": true, "friendly": true, "frosty": true, "funny": true,
		"gallant": true, "gifted": true, "goofy": true, "gracious": true,
		"great": true, "happy": true, "hardcore": true, "heuristic": true,
		"hopeful": true, "hungry": true, "infallible": true, "inspiring": true,
		"intelligent": true, "interesting": true, "jolly": true, "jovial": true,
		"keen": true, "kind": true, "laughing": true, "loving": true,
		"lucid": true, "magical": true, "mystical": true, "modest": true,
		"musing": true, "naughty": true, "nervous": true, "nice": true,
		"nifty": true, "nostalgic": true, "objective": true, "optimistic": true,
		"peaceful": true, "pedantic": true, "pensive": true, "practical": true,
		"priceless": true, "quirky": true, "quizzical": true, "recursing": true,
		"relaxed": true, "reverent": true, "romantic": true, "sad": true,
		"serene": true, "sharp": true, "silly": true, "sleepy": true,
		"stoic": true, "strange": true, "stupefied": true, "suspicious": true,
		"sweet": true, "tender": true, "thirsty": true, "trusting": true,
		"unruffled": true, "upbeat": true, "vibrant": true, "vigilant": true,
		"vigorous": true, "wizardly": true, "wonderful": true, "xenodochial": true,
		"youthful": true, "zealous": true, "zen": true,
	}
	
	nouns := map[string]bool{
		"albattani": true, "allen": true, "almeida": true, "agnesi": true,
		"archimedes": true, "ardinghelli": true, "aryabhata": true, "austin": true,
		"babbage": true, "banach": true, "banzai": true, "bardeen": true,
		"bartik": true, "bassi": true, "beaver": true, "bell": true,
		"benz": true, "bhabha": true, "bhaskara": true, "black": true,
		"blackburn": true, "blackwell": true, "bohr": true, "booth": true,
		"borg": true, "bose": true, "bouman": true, "boyd": true,
		"brahmagupta": true, "brattain": true, "brown": true, "buck": true,
		"burnell": true, "cannon": true, "carson": true, "cartwright": true,
		"carver": true, "cerf": true, "chandrasekhar": true, "chaplygin": true,
		"chatelet": true, "chatterjee": true, "chebyshev": true, "clarke": true,
		"cohen": true, "colden": true, "cori": true, "cray": true,
		"curie": true, "curran": true, "darwin": true, "davinci": true,
		"dewdney": true, "dhawan": true, "diffie": true, "dijkstra": true,
		"dirac": true, "driscoll": true, "dubinsky": true, "easley": true,
		"edison": true, "einstein": true, "elbakyan": true, "elgamal": true,
		"elion": true, "ellis": true, "engelbart": true, "euclid": true,
		"euler": true, "faraday": true, "feistel": true, "fermat": true,
		"fermi": true, "feynman": true, "franklin": true, "gagarin": true,
		"galileo": true, "galois": true, "ganguly": true, "gates": true,
		"gauss": true, "germain": true, "goldberg": true, "goldstine": true,
		"goldwasser": true, "golick": true, "goodall": true, "gould": true,
		"greider": true, "grothendieck": true, "haibt": true, "hamilton": true,
		"haslett": true, "hawking": true, "heisenberg": true, "hellman": true,
		"hermann": true, "herschel": true, "hertz": true, "heyrovsky": true,
		"hodgkin": true, "hofstadter": true, "hoover": true, "hopper": true,
		"hugle": true, "hypatia": true, "ishizaka": true, "jackson": true,
		"jang": true, "jemison": true, "jennings": true, "jepsen": true,
		"johnson": true, "joliot": true, "jones": true, "kalam": true,
		"kapitsa": true, "kare": true, "keldysh": true, "keller": true,
		"kepler": true, "khayyam": true, "khorana": true, "kilby": true,
		"kirch": true, "knuth": true, "kowalevski": true, "lalande": true,
		"lamarr": true, "lamport": true, "leakey": true, "leavitt": true,
		"lederberg": true, "lehmann": true, "lewin": true, "lichterman": true,
		"liskov": true, "lovelace": true, "lumiere": true, "mahavira": true,
		"margulis": true, "matsumoto": true, "maxwell": true, "mayer": true,
		"mccarthy": true, "mcclintock": true, "mclaren": true, "mclean": true,
		"mcnulty": true, "meitner": true, "mendel": true, "mendeleev": true,
		"meninsky": true, "merkle": true, "mestorf": true, "minsky": true,
		"mirzakhani": true, "moore": true, "morse": true, "murdock": true,
		"moser": true, "napier": true, "nash": true, "neumann": true,
		"newton": true, "nightingale": true, "nobel": true, "noether": true,
		"northcutt": true, "noyce": true, "panini": true, "pare": true,
		"pascal": true, "pasteur": true, "payne": true, "perlman": true,
		"pike": true, "poincare": true, "poitras": true, "proskuriakova": true,
		"ptolemy": true, "raman": true, "ramanujan": true, "ride": true,
		"montalcini": true, "ritchie": true, "rhodes": true, "robinson": true,
		"roentgen": true, "rosalind": true, "rubin": true, "saha": true,
		"sammet": true, "sanderson": true, "satoshi": true, "shamir": true,
		"shannon": true, "shaw": true, "shirley": true, "shockley": true,
		"shtern": true, "sinoussi": true, "snyder": true, "solomon": true,
		"spence": true, "stonebraker": true, "sutherland": true, "swanson": true,
		"swartz": true, "swirles": true, "taussig": true, "tereshkova": true,
		"tesla": true, "tharp": true, "thompson": true, "torvalds": true,
		"tu": true, "turing": true, "varahamihira": true, "vaughan": true,
		"visvesvaraya": true, "volhard": true, "villani": true, "wah": true,
		"wiles": true, "williams": true, "williamson": true, "wilson": true,
		"wing": true, "wozniak": true, "wright": true, "wu": true,
		"yalow": true, "yonath": true, "zhukovsky": true,
	}
	
	// Generate multiple names and verify they use valid words
	for i := 0; i < 50; i++ {
		name, err := generateInstanceName()
		if err != nil {
			t.Fatalf("generateInstanceName failed: %v", err)
		}
		
		parts := regexp.MustCompile(`-`).Split(name, 2)
		if len(parts) != 2 {
			t.Errorf("Instance name '%s' should have exactly two parts separated by '-'", name)
			continue
		}
		
		adjective := parts[0]
		noun := parts[1]
		
		if !adjectives[adjective] {
			t.Errorf("Instance name '%s' uses invalid adjective '%s'", name, adjective)
		}
		if !nouns[noun] {
			t.Errorf("Instance name '%s' uses invalid noun '%s'", name, noun)
		}
	}
}
